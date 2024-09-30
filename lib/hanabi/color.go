package hanabi

type Color int

const (
	Red Color = iota
	Yellow
	Green
	Blue
	Purple
	Teal
	Black
	Rainbow
)

var letters = []string{"r", "y", "g", "b", "p", "t", "k", "n"}
var colorNames = []string{
	"red",
	"yellow",
	"green",
	"blue",
	"purple",
	"teal",
	"black",
	"rainbow",
}

func (c Color) Letter() string {
	return letters[c]
}

func (c Color) CanBeClued() bool {
	return c != Rainbow
}

func (c Color) Cards() (cards []CardIdentity) {
	var numbers []int
	if c == Black {
		numbers = []int{1, 2, 3, 4, 5}
	} else {
		numbers = []int{1, 1, 1, 2, 2, 3, 3, 4, 4, 5}
	}
	for _, number := range numbers {
		cards = append(cards, CardIdentity{
			Number: number,
			Color:  c,
		})
	}
	return
}

func (c Color) Name() string {
	return colorNames[c]
}
