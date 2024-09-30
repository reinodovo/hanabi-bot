package main

import (
	"os"

	"github.com/reinodovo/hanabi-bot/lib/client"
)

var password = os.Getenv("BOT_PASSWORD")

type Spawner struct {
	strategy SpawnerStrategy
	tables   map[int]client.Table
}

func findTable(player string, tables map[int]client.Table) *client.Table {
	for _, table := range tables {
		for _, p := range table.Players {
			if p == player {
				return &table
			}
		}
	}
	return nil
}

func (spawner Spawner) handleCommand(chat client.ChatCommand) {
	if chat.Command == "join" {
		pass := ""
		if len(chat.Args) > 0 {
			pass = chat.Args[0]
		}
		table := findTable(chat.Sender, spawner.tables)
		if table == nil {
			return
		}
		go spawner.strategy.Spawn(table.Id, pass)
	} else if chat.Command == "fill" {
		pass := ""
		if len(chat.Args) > 0 {
			pass = chat.Args[0]
		}
		table := findTable(chat.Sender, spawner.tables)
		if table == nil {
			return
		}
		for i := 0; i < int(table.MaxPlayers)-len(table.Players); i++ {
			go spawner.strategy.Spawn(table.Id, pass)
		}
	}
}

func main() {
	cred := client.Credentials{
		User: "reinodovo",
		Pass: password,
	}

	c := client.Connect(cred)
	spawner := Spawner{
		strategy: NewLocalSpawnerStrategy(),
		tables:   make(map[int]client.Table),
	}

	for {
		message, err := c.ReadMessage()
		if err != nil {
			continue
		}

		switch msg := message.(type) {
		case []client.Table:
			spawner.tables = make(map[int]client.Table)
			for _, table := range msg {
				spawner.tables[table.Id] = table
			}
		case client.Table:
			spawner.tables[msg.Id] = msg
		case client.TableGone:
			delete(spawner.tables, msg.TableId)
		case client.ChatCommand:
			spawner.handleCommand(msg)
		}
	}
}
