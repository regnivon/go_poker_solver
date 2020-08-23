package solv
/*
import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConstructTree(t *testing.T) {
	result := ConstructTree(20, 0, 0.75, nil, ipRange, oopRange)

	assert.True(t, result.playerNode == 0)
	assert.True(t, len(result.nextNodes) == 1)
	assert.True(t, result.nextNodes[0].GetNext(0).IsTerminal())
}

func TestConstructTreeComplex(t *testing.T) {
	bets := [][][]float64{
		{
			{
				0.5, 1.0,
			},
		},
		{
			{
				0.5, 1.0,
			},
		},
	}

	result := ConstructTree(20, 20, 0.75, bets, ipRange, oopRange)
	assert.True(t, len(result.nextNodes) == 3)

	check := result.GetNext(0).(*GameNode)

	assert.True(t, check.playerNode == 1)
	assert.True(t, len(check.nextNodes) == 3)

	checkbet10 := result.GetNext(1).(*GameNode)

	assert.True(t, len(checkbet10.nextNodes) == 3)
	assert.True(t, checkbet10.playerNode == 1)
	assert.True(t, checkbet10.potSize == 30)

	checkbet10call := checkbet10.GetNext(0)
	checkbet10fold := checkbet10.GetNext(1)

	assert.True(t, checkbet10call.IsTerminal())
	assert.True(t, checkbet10fold.IsTerminal())
	assert.True(t, checkbet10call.(*ShowdownNode).potSize == 40)
	assert.True(t, checkbet10fold.(*TerminalNode).potSize == 20)

	checkbet10raise20 := checkbet10.GetNext(2).(*GameNode)


	assert.False(t, checkbet10raise20.IsTerminal())
	assert.True(t, checkbet10raise20.potSize == 50)

	checkbet10raise20call := checkbet10raise20.GetNext(0).(*ShowdownNode)

	assert.True(t, checkbet10raise20call.potSize == 60)
	assert.True(t, checkbet10raise20call.oopPlayerStack == 0)
	assert.True(t, checkbet10raise20call.ipPlayerStack == 0)
}
*/