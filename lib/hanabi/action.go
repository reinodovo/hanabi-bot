package hanabi

type Action interface{}

type Play struct {
	Player Player
	CardId int
	Card   CardIdentity
}

type Discard struct {
	Player Player
	CardId int
	Card   CardIdentity
}

type Draw struct {
	Player Player
	CardId int
	Card   CardIdentity
}
