package client

import "fmt"

type Encodable interface {
	tag() string
}

type SendAction struct {
	TableId int `json:"tableID"`
	Type    int `json:"type"`
	Target  int `json:"target"`
	//Value   int `json:"value"`
}

func (SendAction) tag() string {
	return "action"
}

type TableJoin struct {
	Id   int    `json:"tableID"`
	Pass string `json:"password"`
}

func (TableJoin) tag() string {
	return "tableJoin"
}

type GetGameInfo struct {
	TableId   int `json:"tableID"`
	InfoIndex int
}

func (c GetGameInfo) tag() string {
	return fmt.Sprintf("getGameInfo%v", c.InfoIndex)
}

type Loaded struct {
	TableId int `json:"tableID"`
}

func (Loaded) tag() string {
	return "loaded"
}
