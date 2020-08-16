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

func TestCheckCardBoardOverlap(t *testing.T) {
	c1 := poker.NewCard("Ac")
	c2 := poker.NewCard("As")
	c3 := poker.NewCard("7c")
	c4 := poker.NewCard("8s")
	board := []poker.Card{poker.NewCard("Ac"), poker.NewCard("7s"), poker.NewCard("5s"),
		poker.NewCard("3d"), poker.NewCard("2h")}

	assert.True(t, checkCardBoardOverlap(c1, board))
	assert.False(t, checkCardBoardOverlap(c2, board))
	assert.False(t, checkCardBoardOverlap(c3, board))
	assert.False(t, checkCardBoardOverlap(c4, board))
}

func TestConstructPossibleNextCards(t *testing.T) {
	board := []poker.Card{poker.NewCard("Ac"), poker.NewCard("7s"), poker.NewCard("5s"),
		poker.NewCard("3d")}
	ans := constructPossibleNextCards(board, 48)
	assert.True(t, len(ans) == 48)
	for index := range ans {
		assert.True(t, ans[index] > 0)
		assert.False(t, checkCardBoardOverlap(ans[index], board))
	}
}