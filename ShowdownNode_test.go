package solv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//populated from the TerminalNode_test
var sn = NewShowdownNode(gameNode)

func TestShowdownNode_GetUtil(t *testing.T) {
	sn.FillHandRankings(ipRange, oopRange)

	fast := sn.GetUtil(traversal, convertRangeToFloatSlice(oopRange), convertRangeToFloatSlice(ipRange))
	slow := sn.ShowdownSlow(traversal, sn.oopRanks, sn.ipRanks, convertRangeToFloatSlice(ipRange))

	assert.InDeltaSlice(t, slow, fast, 0.001)

	traversal.Traverser = 1
	fast = sn.GetUtil(traversal, convertRangeToFloatSlice(ipRange), convertRangeToFloatSlice(oopRange))
	slow = sn.ShowdownSlow(traversal, sn.ipRanks, sn.oopRanks, convertRangeToFloatSlice(oopRange))

	traversal.Traverser = 0
}

