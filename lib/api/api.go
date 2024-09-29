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

	url := url.URL { Scheme: httpProt, Host: host, Path: login }

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

type Table struct {
	Id uint32 `json:"tableID"`
}

type ServerMessage struct {}

func parseMessage(_ *[]byte) (ServerMessage, error) {
	return ServerMessage{}, nil
}

type Action struct {}

func encodeMessage(Action) {

}

type Client struct {
	dialer websocket.Dialer
	url url.URL
	conn *websocket.Conn
	headers http.Header
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

func (self *Client) sendMessageInternal(msgType string, obj *interface{}) {
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

	parts := bytes.SplitN(msg, []byte(" "), 1)

	switch len(parts) {
		case 1:
			return string(parts[0]), nil
		case 2:
			return string(parts[0]), parts[1]
		default:
			panic("Bad server message")
	}

}

func (self *Client) joinTable(t Table) {
	self.sendMessageInternal("tableJoin", t)
}

func (self *Client) ReadMessage() {
	msgType, content := self.readMessageInternal()

	switch msgType {
		case "ipa":
		case "uga":
	}
}

func (*Client) PerformAction(msg Action) {

}

func Connect(c Credentials, table Table) Client {
	loginToken := Login(c);

	url := url.URL { Scheme: wsProt, Host: host, Path: ws }

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Could not create cookie jar")
	}
	jar.SetCookies(&url, loginToken.cookies)

	dialer := websocket.Dialer { Jar: jar }

	headers := http.Header {}
	for _, cookie := range loginToken.cookies {
		headers.Add("Cookie", cookie.String())
	}

	client := Client {
		dialer: dialer,
		url: url,
		conn: nil,
		headers: headers,
	}

	client.connect()
	client.joinTable(table)

	for {}

	return client
}

