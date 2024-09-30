package hanabi

func HandSize(players int) int {
	if players == 2 {
		return 5
	}
	if players == 6 {
		return 3
	}
	return 4
}
