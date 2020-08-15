package solv

import "github.com/chehsunliu/poker"

//Hand is an alias for an int32 array with two members, preferably with the first being the greater number
type Hand [2]poker.Card
//HandToFloatSliceMap is a map of Hand to a slice of floats for that hand
type HandToFloatSliceMap map[Hand][]float64
//HandToFloatMap maps a hand to a float value
type HandToFloatMap map[Hand]float64

//Combo maps a particular hand to the number of combinations
type Combo struct {
	Hand Hand
	Combos float64
}

type Range []Combo

func NewCombo(hand Hand, combos float64) *Combo {
	return &Combo{Hand: hand, Combos: combos}
}

func (h Hand) String() string {
	return h[0].String() + h[1].String()
}

//NewHand return a pointer to a new hand
func NewHand(c1, c2 string) Hand {
	return Hand{
		poker.NewCard(c1),
		poker.NewCard(c2),
	}
}

