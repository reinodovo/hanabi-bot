package hanabi

import (
	"fmt"
	"math/rand"
)

type Game struct {
	Actions       []Action
	CurrentPlayer Player
	Variant       Variant
	Clues         int
	Strikes       int
	PlayGraph     PlayGraph
	Players       int
	Position      Player
	DiscardPile   []CardIdentity
	Deck          int
	Score         int
	Stacks        []int
	Hands         []Hand
	// TODO: markings
}

func (game Game) CanPlay(card CardIdentity) bool {
	return card.Number == game.Stacks[game.Variant.StackIndex(card.Color)]+1
}

func (game Game) StackTop(color Color) CardIdentity {
	return CardIdentity{
		Number: game.Stacks[game.Variant.StackIndex(color)],
		Color:  color,
	}
}

func (game Game) IsCritical(card CardIdentity) bool {
	var seenCards []CardIdentity
	seenCards = append(seenCards, game.DiscardPile...)
	return game.Variant.IsCritical(card, seenCards)
}

func (game Game) GetCard(id int) Card {
	for _, hand := range game.Hands {
		for _, card := range hand {
			if card.Id == id {
				return card
			}
		}
	}
	return Card{Id: -1}
}

func (game *Game) UpdateCard(updatedCard Card) {
	for player, hand := range game.Hands {
		for slot, card := range hand {
			if card.Id == updatedCard.Id {
				game.Hands[player][slot] = updatedCard
				return
			}
		}
	}
}

func (game *Game) ApplyClue(clue Clue) {
	game.Hands[clue.GetTarget()].ApplyClue(clue)
	game.Clues--
}

func (game *Game) ApplyAction(action Action) {
	game.Actions = append(game.Actions, action)
	switch a := action.(type) {
	case Draw:
		game.Hands[a.Player] = append(game.Hands[a.Player], Card{
			Id:       a.CardId,
			Identity: a.Card,
		})
		game.CurrentPlayer = (game.CurrentPlayer + 1) % game.Players
	case Discard:
		game.DiscardPile = append(game.DiscardPile, a.Card)
		newHand := make(Hand, 0)
		for _, card := range game.Hands[a.Player] {
			if card.Id != a.CardId {
				newHand = append(newHand, card)
			}
		}
		game.Hands[a.Player] = newHand
	case Play:
		newHand := make(Hand, 0)
		for _, card := range game.Hands[a.Player] {
			if card.Id != a.CardId {
				newHand = append(newHand, card)
			}
		}
		game.Hands[a.Player] = newHand
	case Clue:
		game.ApplyClue(a)
		game.CurrentPlayer = (game.CurrentPlayer + 1) % game.Players
	}
}

func RandomGame(v Variant, players int) Game {
	handSize := HandSize(players)
	cards := v.Cards()
	uniqueCards := v.UniqueCards()
	game := Game{
		Clues:     8,
		Variant:   v,
		Strikes:   0,
		Players:   players,
		Score:     0,
		Position:  2,
		Deck:      len(cards) - handSize*players,
		Stacks:    make([]int, len(v.StacksColors())),
		PlayGraph: PlayGraph{CardId: -1, Outgoing: make([]PlayGraphEdge, 0)},
	}
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	game.Hands = make([]Hand, players)
	for id, cardIdentity := range cards {
		if id >= handSize*players {
			break
		}
		card := Card{
			Id:       id,
			Identity: cardIdentity,
			Empathy: UnknownCard{
				PossibleCards: uniqueCards,
				InferedCards:  uniqueCards,
			},
			Known: id%players != game.Position,
		}
		game.Hands[id%players] = append(game.Hands[id%players], card)
	}
	return game
}

func InitGame(variant Variant, players int, position int) Game {
	g := Game{
		Variant:       variant,
		Players:       players,
		Hands:         make([]Hand, players),
		Clues:         8,
		Strikes:       0,
		Score:         0,
		CurrentPlayer: 0,
		Position:      position,
		Actions:       make([]Action, 0),
		Stacks:        make([]int, len(variant.StacksColors())),
		PlayGraph:     PlayGraph{CardId: -1, Outgoing: make([]PlayGraphEdge, 0)},
	}
	for i := 0; i < players; i++ {
		g.Hands[i] = make(Hand, 0)
	}
	return g
}

func (g *Game) Print() {
	for _, hand := range g.Hands {
		for _, card := range hand {
			fmt.Printf("%v%v ", card.Identity.Number, card.Identity.Color.Letter())
		}
		fmt.Println()
	}
}
