package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gorilla/websocket"
)

const host = "hanab.live"
const httpProt = "https"
const wsProt = "wss"
const login = "/login"
const ws = "/ws"

const retryCount = 10

var errUnknownMessageType = fmt.Errorf("unknown message type")

type Client struct {
	dialer  websocket.Dialer
	url     url.URL
	conn    *websocket.Conn
	headers http.Header
}

type Credentials struct {
	User string
	Pass string
}

type LoginToken struct {
	cookies []*http.Cookie
}

func Login(c Credentials) LoginToken {
	form := url.Values{}
	form.Add("username", c.User)
	form.Add("password", c.Pass)
	form.Add("version", "bot")

	url := url.URL{Scheme: httpProt, Host: host, Path: login}

	r, err := http.NewRequest(
		"POST",
		url.String(),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Auth Failed (Status: %v)", res.StatusCode))
	}

	if len(res.Cookies()) == 0 {
		panic("Got no cookies back from the server")
	}

	return LoginToken{
		cookies: res.Cookies(),
	}
}

func (self *Client) establishConnection() (err error) {
	self.conn, _, err = websocket.DefaultDialer.Dial(self.url.String(), self.headers)
	return
}

func (self *Client) SendMessage(obj Encodable) {
	time.Sleep(100 * time.Millisecond)
	var msg string
	var err error

	j, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprint("Could not marshal Encodable of type: ", reflect.TypeOf(obj)))
	}

	if string(j) == "{}" {
		msg = obj.tag()
	} else {
		msg = fmt.Sprintf("%v %v", obj.tag(), string(j))
	}

	bytes := []byte(msg)

	err = retry.Do(
		func() error {
			return self.conn.WriteMessage(websocket.TextMessage, bytes)
		},
		retry.Attempts(retryCount),
		retry.OnRetry(func(n uint, err error) {
			self.establishConnection()
		}),
	)
	if err != nil {
		panic(err)
	}
}

func (self *Client) readMessageInternal() (string, []byte, error) {
	msg, err := retry.DoWithData(
		func() ([]byte, error) {
			_, msg, err := self.conn.ReadMessage()
			return msg, err
		},
		retry.Attempts(retryCount),
		retry.OnRetry(func(n uint, err error) {
			self.establishConnection()
		}),
	)
	if err != nil {
		panic(err)
	}

	parts := bytes.SplitN(msg, []byte{' '}, 2)

	switch len(parts) {
	case 1:
		return string(parts[0]), nil, nil
	case 2:
		return string(parts[0]), parts[1], nil
	default:
		return "", nil, fmt.Errorf("Bad server message")
	}

}

func (self *Client) ReadMessage() (interface{}, error) {
	msgType, content, err := self.readMessageInternal()
	if err != nil {
		return nil, err
	}

	switch msgType {
	case "table":
		table := Table{}
		err := json.Unmarshal(content, &table)
		return table, err
	case "tableList":
		tables := []Table{}
		err := json.Unmarshal(content, &tables)
		return tables, err
	case "chat":
		chat := ChatMessage{}
		err := json.Unmarshal(content, &chat)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(chat.Message, "/") {
			parts := strings.Split(chat.Message, " ")
			command := ChatCommand{
				Sender:  chat.Sender,
				Command: parts[0][1:],
				Args:    parts[1:],
			}
			return command, nil
		}
		return chat, nil
	case "gameAction":
		action := GameAction{}
		err := json.Unmarshal(content, &action)
		return action.Action, err
	case "init":
		init := Init{}
		err := json.Unmarshal(content, &init)
		return init, err
	case "gameActionList":
		actionList := GameActionList{}
		err := json.Unmarshal(content, &actionList)
		return actionList.Actions, err
	case "tableStart":
		table := TableStart{}
		err := json.Unmarshal(content, &table)
		return table, err
	case "tableGone":
		table := TableGone{}
		err := json.Unmarshal(content, &table)
		return table, err
	default:
		return nil, fmt.Errorf("%w: %v", errUnknownMessageType, msgType)
	}
}

func Connect(c Credentials) Client {
	loginToken := Login(c)

	url := url.URL{Scheme: wsProt, Host: host, Path: ws}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Could not create cookie jar")
	}
	jar.SetCookies(&url, loginToken.cookies)

	dialer := websocket.Dialer{Jar: jar}

	headers := http.Header{}
	for _, cookie := range loginToken.cookies {
		headers.Add("Cookie", cookie.String())
	}

	client := Client{
		dialer:  dialer,
		url:     url,
		conn:    nil,
		headers: headers,
	}

	err = retry.Do(func() error {
		return client.establishConnection()
	}, retry.Attempts(retryCount))

	if err != nil {
		panic(err)
	}

	return client
}
