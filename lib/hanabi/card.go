package hanabi

import (
	"fmt"
)

type CardIdentity struct {
	Number int
	Color  Color
}

func (c CardIdentity) Equals(o CardIdentity) bool {
	return c.Color == o.Color && c.Number == o.Number
}

func (c CardIdentity) Previous() CardIdentity {
	return CardIdentity{
		Color:  c.Color,
		Number: c.Number - 1,
	}
}

func (c CardIdentity) Next() CardIdentity {
	return CardIdentity{
		Color:  c.Color,
		Number: c.Number + 1,
	}
}

func (c CardIdentity) String() string {
	return fmt.Sprintf("%v%v", c.Number, c.Color.Letter())
}

type UnknownCard struct {
	PossibleCards []CardIdentity
	InferedCards  []CardIdentity
}

func (c *UnknownCard) ApplyClue(clue Clue, touched bool) {
	var filteredPossibleCards []CardIdentity
	for _, possibleCard := range c.PossibleCards {
		if clue.TouchesCard(possibleCard) == touched {
			filteredPossibleCards = append(filteredPossibleCards, possibleCard)
		}
	}
	var filteredInferedCards []CardIdentity
	for _, inferedCard := range c.InferedCards {
		if clue.TouchesCard(inferedCard) == touched {
			filteredInferedCards = append(filteredInferedCards, inferedCard)
		}
	}
	c.PossibleCards = filteredPossibleCards
	c.InferedCards = filteredInferedCards
}

type Card struct {
	Id        int
	Known     bool
	Identity  CardIdentity
	Empathy   UnknownCard
	Clued     bool
	Finessed  bool
	ChopMoved bool
}

func (c Card) CanBeEqualTo(card CardIdentity) bool {
	if c.Known {
		return c.Identity.Equals(card)
	}
	for _, inferedCard := range c.Empathy.InferedCards {
		if inferedCard.Equals(card) {
			return true
		}
	}
	return false
}
