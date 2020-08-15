package solv

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var gameNode = NewGameNode(0, 20, 15, 15)
var node = NewTerminalNode(gameNode)

var oop = "KQs, /50.0/QJs, KK+, /25.0/55, 87s"
var ip = "KQs, /50.0/QJs, KK+, /25.0/55, 87s"

var oopRange = HandsStringToHandRange(oop)
var ipRange = HandsStringToHandRange(ip)

var traversal = NewTraversal(oopRange, ipRange)

func TestTerminalNode_GetUtil(t *testing.T) {
	traversal.Traverser = 0
	fastResult := node.GetUtil(traversal, convertRangeToFloatSlice(oopRange), convertRangeToFloatSlice(ipRange))
	slowResult := node.TraverserUtilSlow(oopRange, ipRange, node.winUtility)

	assert.InDeltaSlice(t, slowResult, fastResult, 0.01, "Should be equal")

	node.playerNode = 1

	fastResult = node.GetUtil(traversal, convertRangeToFloatSlice(oopRange), convertRangeToFloatSlice(ipRange))
	slowResult = node.TraverserUtilSlow(oopRange, ipRange, -node.winUtility)

	assert.InDeltaSlice(t, slowResult, fastResult, 0.01, "Should be equal")

	node.playerNode = 0

	traversal.Traverser = 1

	fastResult = node.GetUtil(traversal, convertRangeToFloatSlice(ipRange), convertRangeToFloatSlice(oopRange))
	slowResult = node.TraverserUtilSlow(ipRange, oopRange, -node.winUtility)

	assert.InDeltaSlice(t, slowResult, fastResult, 0.01, "Should be equal")
}

func TestTerminalNode_IsTerminal(t *testing.T) {
	assert.True(t, node.IsTerminal())
}

