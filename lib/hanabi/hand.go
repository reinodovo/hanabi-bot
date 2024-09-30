package hanabi

type Hand []Card

const NoChop = -1

func (h Hand) Chop() int {
	for i := len(h) - 1; i >= 0; i-- {
		card := h[i]
		if !card.Clued && !card.ChopMoved && !card.Finessed {
			return i
		}
	}
	return NoChop
}

func (h *Hand) ApplyClue(clue Clue) {
	for slot, card := range *h {
		touched := clue.TouchesSlot(slot)
		card.Empathy.ApplyClue(clue, touched)
		card.Clued = card.Clued || touched
		(*h)[slot] = card
	}
}

func (h Hand) Focus(clue Clue) int {
	chop := h.Chop()
	if chop != NoChop && clue.TouchesSlot(chop) {
		return chop
	}
	var (
		uncluedSlots   []int
		cluedSlots     []int
		chopMovedSlots []int
	)
	for _, slot := range clue.GetTouchedSlots() {
		if h[slot].ChopMoved {
			chopMovedSlots = append(chopMovedSlots, slot)
		} else if h[slot].Clued {
			cluedSlots = append(cluedSlots, slot)
		} else if !h[slot].Finessed {
			uncluedSlots = append(uncluedSlots, slot)
		}
	}
	if len(uncluedSlots) != 0 {
		return uncluedSlots[0]
	}
	if len(chopMovedSlots) != 0 {
		return chopMovedSlots[0]
	}
	if len(cluedSlots) != 0 {
		return cluedSlots[0]
	}
	return clue.GetTouchedSlots()[0]
}
