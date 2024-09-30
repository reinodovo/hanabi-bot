package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

const host = "hanab.live"
const httpProt = "https"
const wsProt = "wss"
const login = "/login"
const ws = "/ws"

const retryCount = 10

type Encodable interface {
	tag() string
}

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

type TableJoin struct {
	Id   uint32 `json:"tableID"`
	Pass string `json:"password"`
}

func (TableJoin) tag() string {
	return "tableJoin"
}

type Table struct {
	Id         uint32   `json:"id"`
	Players    []string `json:"players"`
	MaxPlayers uint32   `json:"maxPlayers"`
}

type ChatMessage struct {
	Message string `json:"msg"`
	Sender  string `json:"who"`
}

type ChatCommand struct {
	Sender  string
	Command string
	Args    []string
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

func (self *Client) establishConnection() {
	var err error
	for i := 0; i < retryCount; i++ {
		self.conn, _, err = websocket.DefaultDialer.Dial(self.url.String(), self.headers)
		if err == nil {
			return
		}
	}
	panic(fmt.Sprint("Could not connect: ", err))
}

func (self *Client) SendMessage(obj Encodable) error {
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

	err = self.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		self.establishConnection()
		err = self.conn.WriteMessage(websocket.TextMessage, bytes)
	}
	return err
}

func (self *Client) readMessageInternal() (string, []byte, error) {
	var msg []byte
	var err error

	_, msg, err = self.conn.ReadMessage()
	if err != nil {
		self.establishConnection()
		_, msg, err = self.conn.ReadMessage()
		if err != nil {
			return "", nil, err
		}
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
	default:
		return nil, fmt.Errorf("Unknown message type: %v", msgType)
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

	client.establishConnection()

	return client
}

func ConnectAndJoin(c Credentials, t TableJoin) Client {
	client := Connect(c)
	client.SendMessage(t)
	return client
}
