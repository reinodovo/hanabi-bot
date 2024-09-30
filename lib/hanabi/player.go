package hanabi

type Player = int

const (
	Alice Player = iota
	Bob
	Cathy
	Donald
)

var names = []string{
	"Alice",
	"Bob",
	"Cathy",
	"Donald",
}

func Name(p Player) string {
	return names[p]
}
