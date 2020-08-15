package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveConflicts(t *testing.T) {
	ipHands := "ATs, 97s"
	board := []poker.Card{poker.NewCard("Ac"), poker.NewCard("7s"), poker.NewCard("5s"),
		poker.NewCard("3d"), poker.NewCard("2h")}

	ip := HandsStringToHandRange(ipHands)

	ip = RemoveConflicts(ip, board)

	fmt.Printf("%v\n", len(ip))
	assert.True(t, len(ip) == 6)
}