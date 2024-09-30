package hanabi

import "fmt"

type Clue interface {
	GetGiver() int
	GetTarget() int
	TouchesCard(card CardIdentity) bool
	GetTouchedSlots() []int
	TouchesSlot(slot int) bool
}

type BaseClue struct {
	Giver        int
	Target       int
	TouchedSlots []int
}

func (c BaseClue) GetGiver() int {
	return c.Giver
}

func (c BaseClue) GetTarget() int {
	return c.Target
}

func (c BaseClue) GetTouchedSlots() []int {
	return c.TouchedSlots
}

func (c BaseClue) TouchesSlot(slot int) bool {
	for _, touchedSlot := range c.TouchedSlots {
		if touchedSlot == slot {
			return true
		}
	}
	return false
}

type NumberClue struct {
	BaseClue
	Number int
}

func (clue *NumberClue) TouchesCard(card CardIdentity) bool {
	return card.Number == clue.Number
}

func (clue *NumberClue) String() string {
	return fmt.Sprintf("%v clued %v to %v", Name(clue.GetGiver()), clue.Number, Name(clue.GetTarget()))
}

type ColorClue struct {
	BaseClue
	Color Color
}

func (clue *ColorClue) TouchesCard(card CardIdentity) bool {
	if card.Color == Rainbow {
		return true
	}
	return card.Color == clue.Color
}

func (clue *ColorClue) String() string {
	return fmt.Sprintf("%v clued %v to %v", Name(clue.GetGiver()), clue.Color.Name(), Name(clue.GetTarget()))
}
