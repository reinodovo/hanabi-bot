package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

const host = "hanab.live"
const httpProt = "https"
const wsProt = "wss"
const login = "/login"
const ws = "/ws"

const retryCount = 10

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

type ServerMessage struct{}

type Action struct{}

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

func parseMessage(_ *[]byte) (ServerMessage, error) {
	return ServerMessage{}, nil
}

func encodeMessage(Action) {

}

func (self *Client) connect() {
	var err error
	for i := 0; i < retryCount; i++ {
		self.conn, _, err = websocket.DefaultDialer.Dial(self.url.String(), self.headers)
		if err == nil {
			return
		}
	}
	panic(fmt.Sprint("Could not connect: ", err))
}

func (self *Client) sendMessageInternal(msgType string, obj interface{}) {
	var msg string
	var err error

	if obj != nil {
		j, err := json.Marshal(obj)
		if err != nil {
			panic(err)
		}
		msg = fmt.Sprintf("%v %v", msgType, string(j))
	} else {
		msg = msgType
	}

	fmt.Println(msg)
	bytes := []byte(msg)

	err = self.conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		self.connect()
		err = self.conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			panic(err)
		}
	}
}

func (self *Client) readMessageInternal() (string, []byte) {
	var msg []byte
	var err error

	_, msg, err = self.conn.ReadMessage()
	if err != nil {
		self.connect()
		_, msg, err = self.conn.ReadMessage()
		if err != nil {
			panic(err)
		}
	}

	parts := bytes.SplitN(msg, []byte{' '}, 2)

	switch len(parts) {
	case 1:
		return string(parts[0]), nil
	case 2:
		return string(parts[0]), parts[1]
	default:
		panic("Bad server message")
	}

}

func (self *Client) SendMessage(msg interface{}) error {
	switch m := msg.(type) {
	case TableJoin:
		self.sendMessageInternal("tableJoin", m)
	default:
		return fmt.Errorf("Unknown message type: %v", msg)
	}
	return nil
}

func (self *Client) ReadMessage() (interface{}, error) {
	msgType, content := self.readMessageInternal()

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

func (*Client) PerformAction(msg Action) {

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

	client.connect()

	return client
}

func ConnectAndJoin(c Credentials, t TableJoin) Client {
	client := Connect(c)
	client.SendMessage(t)
	return client
}
