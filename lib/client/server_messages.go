package client

import "github.com/reinodovo/hanabi-bot/lib/hanabi"

type Table struct {
	Id         int      `json:"id"`
	Players    []string `json:"players"`
	MaxPlayers int      `json:"maxPlayers"`
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

type Action struct {
	Type        string `json:"type"`
	PlayerIndex int    `json:"playerIndex"`
	CardId      int    `json:"order"`
	SuitIndex   int    `json:"suitIndex"`
	Number      int    `json:"rank"`
	Failed      bool   `json:"failed"`
}

type GameAction struct {
	Action Action `json:"action"`
}

type Init struct {
	Position int `json:"ourPlayerIndex"`
	Options  struct {
		Variant hanabi.Variant `json:"variantID"`
		Players int            `json:"numPlayers"`
	} `json:"options"`
}

type GameActionList struct {
	Actions []Action `json:"list"`
}

type TableStart struct {
	TableId int `json:"tableID"`
}
