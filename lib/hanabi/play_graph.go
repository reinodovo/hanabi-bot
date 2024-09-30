package hanabi

import (
	"fmt"
)

type PlayGraphEdgeType int

const (
	NormalEdge PlayGraphEdgeType = iota
	BlindEdge
	PromptEdge
)

type PlayGraphEdge struct {
	Type PlayGraphEdgeType
	Node *PlayGraph
}

type PlayGraph struct {
	CardId   int
	Outgoing []PlayGraphEdge
}

const rootCardId = -1

func (node PlayGraph) CanInsertCard(card CardIdentity, game Game) bool {
	// TODO: return false when card is already inserted
	if game.CanPlay(card) {
		return true
	}
	if node.CardId != rootCardId {
		nodeCard := game.GetCard(node.CardId)
		if nodeCard.CanBeEqualTo(card.Previous()) {
			return true
		}
	}
	canInsertInAnyChild := false
	for _, edge := range node.Outgoing {
		canInsertInAnyChild = edge.Node.CanInsertCard(card, game)
		if canInsertInAnyChild {
			break
		}
	}
	return canInsertInAnyChild
}

func (node *PlayGraph) Find(id int) (*PlayGraph, bool) {
	if node.CardId == id {
		return node, true
	}
	for _, edge := range node.Outgoing {
		n, found := edge.Node.Find(id)
		if found {
			return n, found
		}
	}
	return nil, false
}

func (node *PlayGraph) FindPossibleParents(card CardIdentity, game Game) (parents []*PlayGraph) {
	if node.CardId == rootCardId && game.CanPlay(card) {
		parents = append(parents, node)
		return
	}
	if node.CardId != rootCardId {
		nodeCard := game.GetCard(node.CardId)
		if nodeCard.CanBeEqualTo(card.Previous()) {
			parents = append(parents, node)
		}
	}
	for _, edge := range node.Outgoing {
		parents = append(parents, edge.Node.FindPossibleParents(card, game)...)
	}
	return
}

func (node *PlayGraph) InsertCard(card Card, edgeType PlayGraphEdgeType, game Game) bool {
	if _, alreadyInserted := node.Find(card.Id); alreadyInserted {
		return false
	}
	seenParents := make(map[int]bool)
	finalParents := make([]*PlayGraph, 0)
	for _, inferedCard := range card.Empathy.InferedCards {
		parents := node.FindPossibleParents(inferedCard, game)
		for _, parent := range parents {
			if _, seen := seenParents[parent.CardId]; !seen {
				seenParents[parent.CardId] = true
				finalParents = append(finalParents, parent)
			}
		}
	}
	newNode := PlayGraph{
		CardId:   card.Id,
		Outgoing: make([]PlayGraphEdge, 0),
	}
	edge := PlayGraphEdge{
		Type: edgeType,
		Node: &newNode,
	}
	for _, parent := range finalParents {
		parent.Outgoing = append(parent.Outgoing, edge)
	}
	return true
}

func (node *PlayGraph) PrintMermaidFlowchart(game Game) {
	if node.CardId == rootCardId {
		fmt.Println("flowchart TD")
	}
	for _, edge := range node.Outgoing {
		child := edge.Node
		childCard := game.GetCard(child.CardId)
		fmt.Printf("%v --> %v%v\n", node.CardId, child.CardId, childCard.Empathy.InferedCards)
		child.PrintMermaidFlowchart(game)
	}
}
