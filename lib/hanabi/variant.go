package hanabi

type Variant int

const (
	NormalVariant Variant = iota
	SixSuitsVariant
	BlackVariant
	RainbowVariant
)

func (v Variant) StacksColors() []Color {
	switch v {
	case NormalVariant:
		return []Color{Red, Yellow, Green, Blue, Purple}
	case SixSuitsVariant:
		return []Color{Red, Yellow, Green, Blue, Purple, Teal}
	case RainbowVariant:
		return []Color{Red, Yellow, Green, Blue, Purple, Rainbow}
	case BlackVariant:
		return []Color{Red, Yellow, Green, Blue, Purple, Black}
	default:
		return []Color{}
	}
}

func (v Variant) StackIndex(color Color) int {
	colors := v.StacksColors()
	for i, stackColor := range colors {
		if stackColor == color {
			return i
		}
	}
	return -1
}

func (v Variant) IsCritical(card CardIdentity, seenCards []CardIdentity) bool {
	if card.Color == Black || card.Number == 5 {
		return true
	}
	seenCount := 0
	for _, seenCard := range seenCards {
		if seenCard.Equals(card) {
			seenCount++
		}
	}
	return (seenCount == 2 && card.Number == 1) || seenCount == 1
}

func (v Variant) Cards() (cards []CardIdentity) {
	colors := v.StacksColors()
	for _, color := range colors {
		colorCards := color.Cards()
		cards = append(cards, colorCards...)
	}
	return
}

func (v Variant) UniqueCards() (cards []CardIdentity) {
	colors := v.StacksColors()
	for _, color := range colors {
		for number := 1; number <= 5; number++ {
			cards = append(cards, CardIdentity{Number: number, Color: color})
		}
	}
	return
}
