package solv

import (
	"fmt"
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

func ConstructTree(startingPot, startingStack, defBet float64, bets [][][]float64, ipHands, oopHands Range) *GameNode {
	root := NewGameNode(0, startingPot, startingStack, startingStack)
	addSuccessorNodes(root, 0, bets, defBet)
	initializeNodeHandSlices(root, ipHands, oopHands)
	return root
}

func OutputTree(root Node, level int) {

	if !root.IsTerminal() {
		fmt.Printf("player %v Node pot size %v oop: %v ip: %v\n", root.(*GameNode).playerNode, root.(*GameNode).potSize, root.(*GameNode).oopPlayerStack, root.(*GameNode).ipPlayerStack)

		for i := range root.(*GameNode).nextNodes {
			for i := 0; i < level; i++ {
				fmt.Print("\t")
			}
			node := root.(*GameNode).nextNodes[i]
			fmt.Printf("Action %v ", i)
			OutputTree(node, level+1)
		}
	} else {
		for i := 0; i < level; i++ {
			fmt.Print("\t")
		}
		root.PrintNodeDetails()
	}
}

func addSuccessorNodes(root *GameNode, betNumber int, bets [][][]float64, defBet float64) {
	//b/c, x/x and b/f lines
	if root.playerNode == 1 || betNumber > 0 {
		createNextCallCheckAndFoldNodes(root, betNumber)
	} else {
		//x line
		createCheckToIPNode(root, bets, defBet)
	}
	if root.oopPlayerStack > 0 && root.ipPlayerStack > 0 {
		createNextBetNodes(root, betNumber, bets, defBet)
	}
}

func createNextCallCheckAndFoldNodes(root *GameNode, betNumber int) {
	lastBetSize := math.Abs(root.ipPlayerStack - root.oopPlayerStack)
	callStacks := math.Min(root.ipPlayerStack, root.oopPlayerStack)
	next := NewShowdownNode(NewGameNode(root.playerNode ^ 1, root.potSize + lastBetSize, callStacks, callStacks))
	root.AddNextNode(next)
	if betNumber > 0 {
		foldStacks := math.Max(root.ipPlayerStack, root.oopPlayerStack)
		fold := NewTerminalNode(NewGameNode(root.playerNode ^ 1, root.potSize - lastBetSize, foldStacks, foldStacks))
		root.AddNextNode(fold)
	}
}

func createCheckToIPNode(root *GameNode, bets [][][]float64, defBet float64) {
	gn := NewGameNode(root.playerNode ^ 1, root.potSize, root.ipPlayerStack, root.oopPlayerStack)
	root.AddNextNode(gn)
	addSuccessorNodes(gn, 0, bets, defBet)
}

func createNextBetNodes(root *GameNode, betNumber int, bets [][][]float64, defBet float64) {
	var currentBets []float64

	if bets != nil && len(bets[root.playerNode]) > betNumber {
		currentBets = bets[root.playerNode][betNumber]
	} else {
		currentBets = append(currentBets, defBet)
	}

	//all bet lines
	for index := range currentBets {
		lastBet := math.Abs(root.ipPlayerStack - root.oopPlayerStack)
		sizing := currentBets[index] * (root.potSize + lastBet) + lastBet

		var betSize float64
		var next *GameNode
		if root.playerNode == 1 {
			betSize = math.Min(math.Min(root.ipPlayerStack, sizing), root.oopPlayerStack  + lastBet)
			fmt.Printf("sizing: %v\n", betSize)
			next = NewGameNode(root.playerNode ^ 1, root.potSize + betSize,
				root.ipPlayerStack - betSize, root.oopPlayerStack)
		} else {
			betSize = math.Min(math.Min(root.oopPlayerStack, sizing), root.ipPlayerStack + lastBet)
			fmt.Printf("sizing: %v\n", betSize)
			next = NewGameNode(root.playerNode ^ 1, root.potSize + betSize,
				root.ipPlayerStack, root.oopPlayerStack - betSize)
		}

		root.AddNextNode(next)
		addSuccessorNodes(next, betNumber + 1, bets, defBet)

		if betSize < sizing {
			break
		}
	}
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
		if node, ok := toInit.nextNodes[index].(*ShowdownNode); ok {
			node.FillHandRankings(ipHands, oopHands)
		}
	}
}

func pruneStrategiesAndRegret(root *GameNode) {
	root.resetStrategySums()
	for index := range root.nextNodes {
		if node, ok := root.nextNodes[index].(*GameNode); ok {
			pruneStrategiesAndRegret(node)
		}
	}
}

func Train(traversal *Traversal, iterations int, treeRoot *GameNode) {
	ip := convertRangeToFloatSlice(traversal.Ranges[1])
	oop := convertRangeToFloatSlice(traversal.Ranges[0])
	ipRelativeProb := RangeRelativeProbabilities(traversal.Ranges[1], traversal.Ranges[0])
	oopRelativeProb := RangeRelativeProbabilities(traversal.Ranges[0], traversal.Ranges[1])

	var oopUtil []float64
	var ipUtil []float64
	for i := 0; i <= iterations; i++ {
		traversal.Iteration = i
		traversal.Traverser = 0
		oopUtil = treeRoot.CFRTraversal(traversal, oop, ip)
		traversal.Traverser = 1
		ipUtil = treeRoot.CFRTraversal(traversal, ip, oop)
		if i > 0 && i % 1000 == 0 {
			traversal.Traverser = 0
			oopBestResponse := treeRoot.OverallBestResponse(traversal, oopRelativeProb)
			traversal.Traverser = 1
			ipBestResponse := treeRoot.OverallBestResponse(traversal, ipRelativeProb)
			fmt.Printf("Iteration %v oop BR: %v ip BR: %v exploitability = ", i, oopBestResponse, ipBestResponse)
			fmt.Printf("%v percent of the pot\n", (oopBestResponse + ipBestResponse) / 2 / treeRoot.potSize * 100)
			//pruneStrategiesAndRegret(treeRoot)
		}
	}
	fmt.Println(len(oopUtil))
	fmt.Println(len(ipUtil))
}