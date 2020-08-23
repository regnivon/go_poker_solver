package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
	"math"
)

//Traversal contains the index of the current traverser, the ranges, and two caches mapping
//a hand to the index in the opponents range. This cache is used for terminal node utility eval
type Traversal struct {
	Traverser int
	Ranges [2]Range
	IndexCaches [2]map[Hand]int
	Iteration int
	alpha float64
	beta float64
	gamma float64
}

//ConstructionParams - used for construction of the game tree, the i-th index of a given bets array gives that
//bet number i.e. oopFlopBets[1] gives a slice with the bets used when responding to a one bet sequence prior
//allInCutoff will make the only bet all in if the bettor's stack is smaller than that % of the pot.
//The default bet is a % of the pot used when there are no specific bets for that action sequence.
type ConstructionParams struct {
	allInCutoff float64
	defaultBet float64
	ipFlopBets [][]float64
	oopFlopBets [][]float64
	ipTurnBets [][]float64
	oopTurnBets [][]float64
	ipRiverBets [][]float64
	oopRiverBets [][]float64
}

func NewTraversal(oopRange, ipRange Range) *Traversal {
	var rng [2]Range
	var indexes [2]map[Hand]int
	indexes[0] = make(map[Hand]int)
	indexes[1] = make(map[Hand]int)
	rng[0] = oopRange
	rng[1] = ipRange
	for index := range oopRange {
		indexes[0][oopRange[index].Hand] = index
	}
	for index := range ipRange {
		indexes[1][ipRange[index].Hand] = index
	}
	return &Traversal{
		Traverser: 0,
		Ranges:    rng,
		IndexCaches: indexes,
		Iteration: 0,
		alpha: 1.5,
		beta: 0.0,
		gamma: 2.0,
	}
}

func (traversal* Traversal) GetRange(player int) Range {
	return traversal.Ranges[player]
}

func NewConstructionParams(defaultBet, allInCutoff float64) *ConstructionParams {
	return &ConstructionParams{
		defaultBet: defaultBet,
		allInCutoff: allInCutoff,
	}
}

func ConstructTree(startingPot, startingStack float64, params *ConstructionParams,
					ipHands, oopHands Range, board []poker.Card) *GameNode {
	root := NewGameNode(0, startingPot, startingStack, startingStack)
	cache := NewRiverEvaluationCache(oopHands, ipHands)
	addSuccessorNodes(root, 0, params, board, cache)
	initializeNodeHandSlices(root, ipHands, oopHands)
	return root
}

func OutputTree(root Node) {
	root.PrintNodeDetails(0)
}

func addSuccessorNodes(root *GameNode, betNumber int, params *ConstructionParams,
						board []poker.Card, cache *RiverEvaluationCache) {
	var street int
	switch len(board) {
	case 3:
		street = 1
	case 4:
		street = 2
	case 5:
		street = 3
	}
	//b/c, x/x and b/f lines
	if root.playerNode == 1 || betNumber > 0 {
		createNextCallCheckAndFoldNodes(root, betNumber, street, params, board, cache)
	} else {
		//x line
		createCheckToIPNode(root, params, board, cache)
	}
	if root.oopPlayerStack > 0 && root.ipPlayerStack > 0 {
		createNextBetNodes(root, betNumber, street, params, board, cache)
	}
}

func createNextCallCheckAndFoldNodes(root *GameNode, betNumber, street int, params *ConstructionParams,
									 board []poker.Card, cache *RiverEvaluationCache) {
	lastBetSize := math.Abs(root.ipPlayerStack - root.oopPlayerStack)
	callStacks := math.Min(root.ipPlayerStack, root.oopPlayerStack)
	//go to showdown if this is the river
	if street == 3 {
		next := NewShowdownNode(root.potSize + lastBetSize, root.playerNode, board, cache)
		root.AddNextNode(next)
		index := cache.InsertBoard(board)
		next.cacheIndex = index
	} else if callStacks == 0 {
		next := NewAllInShowdownNode(root.potSize + lastBetSize, street, cache)
		runouts := constructPossibleRunouts(board, cache)
		for _, runout := range runouts {
			showdown := NewShowdownNode(next.potSize, root.playerNode, runout, cache)
			index := cache.InsertBoard(runout)
			showdown.cacheIndex = index
			next.AddNextNode(showdown)
		}
		root.AddNextNode(next)
	} else {
		next := NewChanceNode(root.potSize + lastBetSize, callStacks, board, street)
		for _, card := range next.nextCards {
			newBoard := make([]poker.Card, len(board))
			copy(newBoard, board)
			newBoard = append(newBoard, card)
			gn := NewGameNode(0, next.potSize, next.ipPlayerStack, next.oopPlayerStack)
			next.AddNextNode(gn)
			addSuccessorNodes(gn, 0, params, newBoard, cache)
		}
		root.AddNextNode(next)
	}
	if betNumber > 0 {
		foldStacks := math.Max(root.ipPlayerStack, root.oopPlayerStack)
		fold := NewTerminalNode(NewGameNode(root.playerNode ^ 1, root.potSize - lastBetSize, foldStacks, foldStacks))
		fold.board = board
		root.AddNextNode(fold)
	}
}

func createCheckToIPNode(root *GameNode, params *ConstructionParams, board []poker.Card, cache *RiverEvaluationCache) {
	gn := NewGameNode(root.playerNode ^ 1, root.potSize, root.ipPlayerStack, root.oopPlayerStack)
	root.AddNextNode(gn)
	addSuccessorNodes(gn, 0, params, board, cache)
}

func createNextBetNodes(root *GameNode, betNumber, street int, params *ConstructionParams,
						board []poker.Card, cache *RiverEvaluationCache) {

	currentBets := getCurrentBets(street, root.playerNode, betNumber, params)

	if root.potSize * params.allInCutoff >= math.Max(root.ipPlayerStack, root.oopPlayerStack) {
		currentBets = []float64{params.allInCutoff}
	}
	//all bet lines
	for index := range currentBets {
		lastBet := math.Abs(root.ipPlayerStack - root.oopPlayerStack)
		sizing := currentBets[index] * (root.potSize + lastBet) + lastBet

		var betSize float64
		var next *GameNode
		if root.playerNode == 1 {
			betSize = math.Min(math.Min(root.ipPlayerStack, sizing), root.oopPlayerStack  + lastBet)
			next = NewGameNode(root.playerNode ^ 1, root.potSize + betSize,
				root.ipPlayerStack - betSize, root.oopPlayerStack)
		} else {
			betSize = math.Min(math.Min(root.oopPlayerStack, sizing), root.ipPlayerStack + lastBet)
			next = NewGameNode(root.playerNode ^ 1, root.potSize + betSize,
				root.ipPlayerStack, root.oopPlayerStack - betSize)
		}

		root.AddNextNode(next)
		addSuccessorNodes(next, betNumber + 1, params, board, cache)

		if betSize < sizing {
			break
		}
	}
}

func getCurrentBets(street, player, betNumber int, params *ConstructionParams) []float64 {
	switch street {
	case 1:
		if player == 0 && betNumber < len(params.oopFlopBets) {
			return params.oopFlopBets[betNumber]
		} else if betNumber < len(params.ipFlopBets){
			return params.ipFlopBets[betNumber]
		}
	case 2:
		if player == 0 && betNumber < len(params.oopTurnBets) {
			return params.oopTurnBets[betNumber]
		} else if betNumber < len(params.ipTurnBets){
			return params.ipTurnBets[betNumber]
		}
	case 3:
		if player == 0 && betNumber < len(params.oopRiverBets) {
			return params.oopRiverBets[betNumber]
		} else if betNumber < len(params.ipRiverBets){
			return params.ipRiverBets[betNumber]
		}
	}
	return []float64{params.defaultBet}
}

func initializeNodeHandSlices(toInit *GameNode, ipHands, oopHands Range) {
	toInit.numActions = len(toInit.nextNodes)
	if toInit.playerNode == 1 {
		toInit.InitializeHandSlices(len(ipHands))
	} else {
		toInit.InitializeHandSlices(len(oopHands))
	}
	for index := range toInit.nextNodes {
		if node, ok := toInit.nextNodes[index].(*GameNode); ok {
			initializeNodeHandSlices(node, ipHands, oopHands)
		}
		if node, ok := toInit.nextNodes[index].(*ChanceNode); ok {
			for chanceNextIndex := range node.nextNodes {
				if nextNode, ok := node.nextNodes[chanceNextIndex].(*GameNode); ok {
					initializeNodeHandSlices(nextNode, ipHands, oopHands)
				}
			}
		}
	}
}

func Train(traversal *Traversal, iterations int, treeRoot *GameNode) {
	ip := convertRangeToFloatSlice(traversal.Ranges[1])
	oop := convertRangeToFloatSlice(traversal.Ranges[0])
	ipRelativeProb := RangeRelativeProbabilities(traversal.Ranges[1], traversal.Ranges[0])
	oopRelativeProb := RangeRelativeProbabilities(traversal.Ranges[0], traversal.Ranges[1])

	traversal.Traverser = 0
	oopBestResponse := treeRoot.OverallBestResponse(traversal, oopRelativeProb)
	traversal.Traverser = 1
	ipBestResponse := treeRoot.OverallBestResponse(traversal, ipRelativeProb)

	fmt.Printf("Iteration 0 oop BR: %v ip BR: %v exploitability = ", oopBestResponse, ipBestResponse)
	fmt.Printf("%v percent of the pot\n", (oopBestResponse + ipBestResponse) / 2 / treeRoot.potSize * 100)

	for i := 0; i <= iterations; i++ {
		traversal.Iteration = i
		traversal.Traverser = 0
		treeRoot.CFRTraversal(traversal, oop, ip)
		traversal.Traverser = 1
		treeRoot.CFRTraversal(traversal, ip, oop)
		if  i > 0 && i % 25 == 0 {
			traversal.Traverser = 0
			oopBestResponse := treeRoot.OverallBestResponse(traversal, oopRelativeProb)
			traversal.Traverser = 1
			ipBestResponse := treeRoot.OverallBestResponse(traversal, ipRelativeProb)
			fmt.Printf("Iteration %v oop BR: %v ip BR: %v exploitability = ", i, oopBestResponse, ipBestResponse)
			fmt.Printf("%v percent of the pot\n", (oopBestResponse + ipBestResponse) / 2 / treeRoot.potSize * 100)
		}
	}
}