package main

import (
	"fmt"
	"os"

	"github.com/reinodovo/hanabi-bot/lib/api"
)

var botCount = 0
var password = os.Getenv("PASSWORD")

func findTable(player string, tables map[uint32]api.Table) *api.Table {
	for _, table := range tables {
		for _, p := range table.Players {
			if p == player {
				return &table
			}
		}
	}
	return nil
}

func spawnBot(table api.Table, pass string) {
	bot := api.Credentials{
		User: fmt.Sprintf("ovo-test-%v", botCount),
		Pass: password,
	}
	botCount++
	go func() {
		api.ConnectAndJoin(bot, api.TableJoin{
			Id:   table.Id,
			Pass: pass,
		})
	}()
}

func main() {
	cred := api.Credentials{
		User: "reinodovo",
		Pass: password,
	}

	client := api.Connect(cred)

	tables := make(map[uint32]api.Table)

	for {
		message, err := client.ReadMessage()
		if err != nil {
			//log.Println(err)
		}
		switch msg := message.(type) {
		case []api.Table:
			tables = make(map[uint32]api.Table)
			for _, table := range msg {
				tables[table.Id] = table
			}
			fmt.Println(len(msg))
		case api.Table:
			tables[msg.Id] = msg
			fmt.Println(len(tables))
		case api.ChatCommand:
			if msg.Command == "join" {
				pass := ""
				if len(msg.Args) > 0 {
					pass = msg.Args[0]
				}
				table := findTable(msg.Sender, tables)
				if table == nil {
					continue
				}
				spawnBot(*table, pass)
			} else if msg.Command == "fill" {
				pass := ""
				if len(msg.Args) > 0 {
					pass = msg.Args[0]
				}
				table := findTable(msg.Sender, tables)
				if table == nil {
					continue
				}
				for i := 0; i < int(table.MaxPlayers)-len(table.Players); i++ {
					spawnBot(*table, pass)
				}
			}
		}
	}
}
