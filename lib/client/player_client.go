package client

import (
	"fmt"

	"github.com/reinodovo/hanabi-bot/lib/hanabi"
)

type PlayerClient struct {
	client      Client
	tableId     int
	actionCount int
	variant     hanabi.Variant
}

func ConnectAndJoin(c Credentials, t TableJoin) PlayerClient {
	client := Connect(c)
	client.SendMessage(t)
	return PlayerClient{
		client:  client,
		tableId: t.Id,
	}
}

func (c *PlayerClient) PerformAction(action hanabi.Action) {
	switch a := action.(type) {
	case hanabi.Play:
		c.client.SendMessage(SendAction{
			TableId: c.tableId,
			Type:    0,
			Target:  a.CardId,
		})
	case hanabi.Discard:
		c.client.SendMessage(SendAction{
			TableId: c.tableId,
			Type:    1,
			Target:  a.CardId,
		})
	}
}

func (c *PlayerClient) parseAction(action Action) (hanabi.Action, error) {
	if action.SuitIndex == -1 {
		action.SuitIndex = 0
	}
	switch action.Type {
	case "discard":
		return hanabi.Discard{
			Player: action.PlayerIndex,
			CardId: action.CardId,
			Card: hanabi.CardIdentity{
				Color:  c.variant.StacksColors()[action.SuitIndex],
				Number: action.Number,
			},
		}, nil
	case "draw":
		return hanabi.Draw{
			Player: action.PlayerIndex,
			CardId: action.CardId,
			Card: hanabi.CardIdentity{
				Color:  c.variant.StacksColors()[action.SuitIndex],
				Number: action.Number,
			},
		}, nil
	case "play":
		return hanabi.Play{
			Player: action.PlayerIndex,
			CardId: action.CardId,
			Card: hanabi.CardIdentity{
				Color:  c.variant.StacksColors()[action.SuitIndex],
				Number: action.Number,
			},
		}, nil
	default:
		return nil, fmt.Errorf("Unknown action type: %v", action.Type)
	}
}

func (c *PlayerClient) ReadMessage() (interface{}, error) {
	for {
		message, err := c.client.ReadMessage()
		if err != nil {
			continue
		}
		switch msg := message.(type) {
		case TableStart:
			c.actionCount = 0
			c.tableId = msg.TableId
			c.client.SendMessage(GetGameInfo{TableId: c.tableId, InfoIndex: 1})
			if err != nil {
				panic(err)
			}
		case Init:
			c.client.SendMessage(GetGameInfo{TableId: c.tableId, InfoIndex: 2})
			if err != nil {
				panic(err)
			}
			c.variant = msg.Options.Variant
			game := hanabi.InitGame(msg.Options.Variant, msg.Options.Players, msg.Position)
			return game, nil
		case []Action:
			c.client.SendMessage(Loaded{TableId: c.tableId})
			if err != nil {
				panic(err)
			}
			if len(msg) > c.actionCount {
				var pendingActions []hanabi.Action
				for i := c.actionCount; i < len(msg); i++ {
					action, err := c.parseAction(msg[i])
					if err != nil {
						continue
					}
					pendingActions = append(pendingActions, action)
				}
				c.actionCount = len(msg)
				return pendingActions, nil
			}
		case Action:
			c.actionCount++
			action, err := c.parseAction(msg)
			if err != nil {
				continue
			}
			return []hanabi.Action{action}, nil
		default:
			continue
		}
	}
}
