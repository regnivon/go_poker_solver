package solv

import (
	"fmt"
	"math"
)

type Node interface {
	CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64
	BestResponse(traversal *Traversal, opponentReachProb []float64) []float64
	PrintNodeDetails(level int)
}

type GameNode struct {
	playerNode int
	numActions int
	isTerminal bool
	potSize float64
	ipPlayerStack float64
	oopPlayerStack float64

	regrets      [][]float64
	strategies   [][]float64
	strategySums [][]float64

	nextNodes []Node
}

func NewGameNode(playerNode int, potSize float64, ipPlayerStack float64, oopPlayerStack float64) *GameNode {
	var node = GameNode{playerNode: playerNode, potSize: potSize, ipPlayerStack: ipPlayerStack,
		oopPlayerStack: oopPlayerStack, numActions: 0}
	node.isTerminal = false
	node.nextNodes = make([]Node, 0)
	return &node
}

//PotSize returns the potsize
func (node *GameNode) PotSize() float64 {
	return node.potSize
}

//IPPlayerNode returns if this is an ip player node
func (node *GameNode) PlayerNode() int {
	return node.playerNode
}

//OOPPlayerStack returns OOP player stack size
func (node *GameNode) OOPPlayerStack() float64 {
	return node.oopPlayerStack
}

//IPPlayerStack returns IP player stack size
func (node *GameNode) IPPlayerStack() float64 {
	return node.ipPlayerStack
}

//StrategySums returns the strategy sum vector for a given hand
func (node *GameNode) StrategySums(hand int) []float64 {
	return node.strategySums[hand]
}

//Regrets returns the regret sum vector for a given hand
func (node *GameNode) Regrets(hand int) []float64 {
	return node.regrets[hand]
}

//Strategy returns the strategy vector for a given hand
func (node *GameNode) Strategy(hand int) []float64 {
	return node.strategies[hand]
}

//IsTerminal returns if this node is terminal
func (node *GameNode) IsTerminal() bool {
	return node.isTerminal
}

func (node *GameNode) NumActions() int {
	return node.numActions
}

func (node *GameNode) AddNextNode(next Node) {
	node.nextNodes = append(node.nextNodes, next)
}

func (node *GameNode) GetNext(index int) Node {
	return node.nextNodes[index]
}

func (node *GameNode) InitializeHandSlices(numberHands int) {
	node.regrets      = make([][]float64, numberHands)
	node.strategies   = make([][]float64, numberHands)
	node.strategySums = make([][]float64, numberHands)
	for i := 0; i < numberHands; i++ {
		node.regrets[i]      = make([]float64, node.numActions)
		node.strategies[i]   = make([]float64, node.numActions)
		node.strategySums[i] = make([]float64, node.numActions)
	}
}

func (node *GameNode) RegretMatchAllHands() {
	for hand := range node.regrets {
		node.regretMinimize(hand)
	}
}

//RegretMinimize performs the strategy update for a given hand with the regret formula
func (node *GameNode) regretMinimize(hand int) {
	normalizingSum := node.regretStrategyUpdate(hand)
	node.NormalizeStrategy(hand, normalizingSum)
}

//RegretStrategyUpdate sets the strategy for a given hand based on the regret values
func (node *GameNode) regretStrategyUpdate(hand int) float64 {
	normalizingSum := 0.0
	regrets := node.Regrets(hand)
	for i := 0; i < node.numActions; i++ {
		node.strategies[hand][i] = math.Max(regrets[i], 0.0)
		normalizingSum += node.strategies[hand][i]
	}
	return normalizingSum
}

//NormalizeStrategy normalizes the strategy vector for a given hand utilizing the normalizing sum
func (node *GameNode) NormalizeStrategy(hand int, normalizingSum float64) {
	for i := 0; i < node.numActions; i++ {
		if normalizingSum > 0 {
			node.strategies[hand][i] /= normalizingSum
		} else {
			node.strategies[hand][i] = 1.0 / float64(node.numActions)
		}
	}
}

func (node *GameNode) RegretAndStrategySumsUpdate(trav *Traversal, reachProbability, nodeUtility []float64, actionUtility [][]float64) {
	iter := float64(trav.Iteration)
	alpha := math.Pow(iter, trav.alpha)
	beta := math.Pow(iter, trav.beta)
	gamma := iter / (iter + 1.0)
	positiveRegret := alpha / (alpha + 1.0)
	negativeRegret := beta / (beta + 1.0)
	strategyWeight := math.Pow(gamma, trav.gamma)

	for hand := range reachProbability {
		for i := 0; i < node.numActions; i++ {
			node.strategySums[hand][i] += reachProbability[hand] * node.strategies[hand][i] * strategyWeight
			node.strategySums[hand][i] *= strategyWeight
			node.regrets[hand][i] += actionUtility[i][hand] - nodeUtility[hand]
			//node.strategySums[hand][i] += reachProbability[hand] * node.strategies[hand][i]
			if node.regrets[hand][i] > 0 {
				node.regrets[hand][i] *= positiveRegret
			} else {
				node.regrets[hand][i] *= negativeRegret
			}
		}
	}
}

func (node *GameNode) resetStrategySums() {
	for hand := range node.strategySums {
		for i := 0; i < node.numActions; i++ {
			node.strategySums[hand][i] = 0.0
			//if node.regrets[hand][i] < 0 {
			//	node.regrets[hand][i] = 0.0
			//}
		}
	}
}

func (node *GameNode) GetAverageStrategy() [][]float64 {
	strategies := make([][]float64, len(node.strategies))
	for hand := range node.strategySums {
		strategies[hand] = node.getAverageStrategy(hand)
	}
	return strategies
}

func (node *GameNode) getAverageStrategy(hand int) []float64 {
	normalizingSum := 0.0
	averageStrategy := make([]float64, node.numActions)
	for i:=0; i < node.numActions; i++ {
		normalizingSum += node.strategySums[hand][i]
	}
	for i:=0; i < node.numActions; i++ {
		if normalizingSum > 0 {
			averageStrategy[i] = node.strategySums[hand][i] / normalizingSum
		} else {
			averageStrategy[i] = 1.0 / float64(node.numActions)
		}
	}
	return averageStrategy
}

func (node *GameNode) PrintNodeDetails(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("GameNode player: %v potsize: %v oop stack: %v ip stack %v\n", node.playerNode, node.potSize, node.oopPlayerStack, node.ipPlayerStack)
	for _, child := range node.nextNodes {
		child.PrintNodeDetails(level + 1)
	}
}

func (node *GameNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {

	node.RegretMatchAllHands()
	nodeUtility := make([]float64, len(traverserReachProb))
	if traversal.Traverser == node.playerNode {
		node.TraverserCFR(traversal, traverserReachProb, opponentReachProb, nodeUtility)
	} else {
		node.OpponentCFR(traversal, traverserReachProb, opponentReachProb, nodeUtility)
	}
	return nodeUtility
}

func (node *GameNode) TraverserCFR(traversal *Traversal, traverserReachProb, opponentReachProb, nodeUtility []float64) {
	storedUtility := make([][]float64, node.numActions)
	for i := range storedUtility {
		storedUtility[i] = make([]float64, len(traverserReachProb))
	}

	for i := 0; i < node.numActions; i++ {
		nextReachProb := make([]float64, len(traverserReachProb))
		for hand := range traverserReachProb {
			nextReachProb[hand] = node.strategies[hand][i] * traverserReachProb[hand]
		}
		result := node.GetNext(i).CFRTraversal(traversal, nextReachProb, opponentReachProb)
		for hand := range result {
			storedUtility[i][hand] = result[hand]
			nodeUtility[hand] += node.strategies[hand][i] * result[hand]
		}
	}

	node.RegretAndStrategySumsUpdate(traversal, traverserReachProb, nodeUtility, storedUtility)
}

func (node *GameNode) OpponentCFR(traversal *Traversal, traverserReachProb, opponentReachProb, nodeUtility []float64) {
	for i := 0; i < node.numActions; i++ {
		nextReachProb := make([]float64, len(opponentReachProb))
		for hand := range opponentReachProb {
			nextReachProb[hand] = node.strategies[hand][i] * opponentReachProb[hand]
		}
		result := node.GetNext(i).CFRTraversal(traversal, traverserReachProb, nextReachProb)
		for hand := range result {
			nodeUtility[hand] += result[hand]
		}
	}
}

//BestResponse traverses the game tree and finds the ev of the best response strategy for the responding player
func (node *GameNode) OverallBestResponse(traversal *Traversal, responderRelativeProbs []float64) float64 {
	responder := traversal.Traverser
	opponent := responder ^ 1

	oppProb := convertRangeToFloatSlice(traversal.Ranges[opponent])
	unblocked := UnblockedHands(traversal.Ranges[responder], traversal.Ranges[opponent])
	evs :=  node.BestResponse(traversal, oppProb)
	sum := 0.0

	for i := range traversal.Ranges[responder] {
		sum += evs[i] * responderRelativeProbs[i] / unblocked[i]
	}
	return sum
}

//BestResponse calculates the best response ev for a specific hand through a recursive tree search
func (node *GameNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	if node.playerNode == traversal.Traverser {
		bestEvs := make([]float64, len(node.strategies))
		for i := range node.nextNodes {
			nextEv := node.nextNodes[i].BestResponse(traversal, opponentReachProb)
			for hand := range bestEvs {
				if i == 0 || nextEv[hand] > bestEvs[hand] {
					bestEvs[hand] = nextEv[hand]
				}
			}
		}
		return bestEvs
	} else {
		nodeEv := make([]float64, len(traversal.Ranges[traversal.Traverser]))
		averageStrategies := make([][]float64, len(node.strategies))

		for i := range opponentReachProb {
			averageStrategies[i] = node.getAverageStrategy(i)
		}

		for i := range node.nextNodes {
			nextReach := make([]float64, len(opponentReachProb))
			for j := range opponentReachProb {
				nextReach[j] = averageStrategies[j][i] * opponentReachProb[j]
			}

			nextEv := node.nextNodes[i].BestResponse(traversal, nextReach)
			for j := range nextEv {
				nodeEv[j] += nextEv[j]
			}
		}
		return nodeEv
	}
}



