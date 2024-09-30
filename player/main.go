package main

import (
	"os"
	"strconv"

	"github.com/reinodovo/hanabi-bot/lib/client"
	"github.com/reinodovo/hanabi-bot/lib/hanabi"
)

func play(player client.PlayerClient) {
	game := hanabi.Game{}
	for {
		message, err := player.ReadMessage()
		if err != nil {
			continue
		}

		switch msg := message.(type) {
		case hanabi.Game:
			game = msg
		case []hanabi.Action:
			for _, action := range msg {
				game.ApplyAction(action)
			}
			if game.CurrentPlayer == game.Position {
				player.PerformAction(hanabi.Play{
					Player: game.Position,
					CardId: game.Hands[game.Position][0].Id,
				})
			}
		}
	}
}

func main() {
	bot := client.Credentials{
		User: os.Getenv("BOT_NAME"),
		Pass: os.Getenv("BOT_PASSWORD"),
	}
	tableId, err := strconv.Atoi(os.Getenv("TABLE_ID"))
	if err != nil {
		panic(err)
	}
	player := client.ConnectAndJoin(bot, client.TableJoin{
		Id:   tableId,
		Pass: os.Getenv("TABLE_PASSWORD"),
	})
	play(player)
}
