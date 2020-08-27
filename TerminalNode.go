package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
)

//TerminalNode is a GameNode that occurs in the GameTree after the prior node action is a fold
//thus, !ipPlayerNode tells you who just folded and thus has a negative EV at this node
type TerminalNode struct {
	*GameNode
	winUtility float64
	board []poker.Card
}

//NewTerminalNode constructs a TerminalNode
func NewTerminalNode(gameNode *GameNode) *TerminalNode {
	node := TerminalNode{GameNode: gameNode}
	node.isTerminal = true
	node.winUtility = node.potSize / 2.0
	return &node
}

func (node *TerminalNode) IsTerminal() bool {
	return node.isTerminal
}

func (node *TerminalNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	return node.GetUtil(traversal, traverserReachProb, opponentReachProb)
}

//GetUtil accepts the if the current traverser is IP, the reach probabilities for each player and then returns
//a map of hand to utility
func (node *TerminalNode) GetUtil(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	var result []float64
	//if the traverser is the same as the terminal node that means they are +ev,
	if traversal.Traverser == node.playerNode {
		result =  node.TraverserUtil(traversal, traverserReachProb, opponentReachProb, node.winUtility)
	} else {
		result =  node.TraverserUtil(traversal, traverserReachProb, opponentReachProb, -node.winUtility)
	}
	return result
}

//TraverserUtil performs a O(N) terminal utility calculation by calculating the total reach probability
//for the opponent at this node and the card removal probability for each card in a single pass
//This information is then used to update the EV of each of our hands
func (node *TerminalNode) TraverserUtil(traversal *Traversal, travProb, oppProb []float64, utility float64) []float64 {
	cardRemoval := make(map[poker.Card]float64)
	utilities := make([]float64, len(travProb))
	probabilitySum := 0.0

	opponent := traversal.Traverser ^ 1
	traverserHands := traversal.GetRange(traversal.Traverser)
	oopHands := traversal.GetRange(opponent)

	for index := range oopHands {
		cardRemoval[oopHands[index].Hand[0]] += oppProb[index]
		cardRemoval[oopHands[index].Hand[1]] += oppProb[index]
		probabilitySum += oppProb[index]
	}

	for index := range traverserHands {
		if CheckHandBoardOverlap(traverserHands[index].Hand, node.board) {
			continue
		}
		var removal float64
		sameHandIndex, ok := traversal.IndexCaches[opponent][traverserHands[index].Hand]
		if ok {
			removal = -cardRemoval[traverserHands[index].Hand[0]] - cardRemoval[traverserHands[index].Hand[1]] +
				      oppProb[sameHandIndex]
		} else {
			removal = -cardRemoval[traverserHands[index].Hand[0]] - cardRemoval[traverserHands[index].Hand[1]]
		}
		utilities[index] = (probabilitySum + removal) * utility
	}
	return utilities
}

//TraverserUtilSlow this is an O(n^2) utility calculation, used for the naive comparison of the O(n) algorithm
func (node *TerminalNode) TraverserUtilSlow(travProb, oppProb Range, utility float64) []float64 {
	UtilityMap := make([]float64, len(travProb))
	for traverserHand := range travProb {
		for opponentHand := range oppProb {
			if !CheckHandOverlap(travProb[traverserHand].Hand, oppProb[opponentHand].Hand) {
				UtilityMap[traverserHand] += oppProb[opponentHand].Combos
			}
		}
		UtilityMap[traverserHand] *= utility
	}
	return UtilityMap
}

func (node *TerminalNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	var utility float64
	if traversal.Traverser == node.playerNode {
		utility = node.winUtility
	} else {
		utility = -node.winUtility
	}
	cardRemoval := make(map[poker.Card]float64)
	utilities := make([]float64, len(traversal.Ranges[traversal.Traverser]))
	probabilitySum := 0.0

	opponent := traversal.Traverser ^ 1
	traverserHands := traversal.GetRange(traversal.Traverser)
	oopHands := traversal.GetRange(opponent)

	for index := range oopHands {
		cardRemoval[oopHands[index].Hand[0]] += opponentReachProb[index]
		cardRemoval[oopHands[index].Hand[1]] += opponentReachProb[index]
		probabilitySum += opponentReachProb[index]
	}

	for index := range traverserHands {
		if CheckHandBoardOverlap(traverserHands[index].Hand, node.board) {
			continue
		}
		var removal float64
		sameHandIndex, ok := traversal.IndexCaches[opponent][traverserHands[index].Hand]
		if ok {
			removal = -cardRemoval[traverserHands[index].Hand[0]] - cardRemoval[traverserHands[index].Hand[1]] +
				opponentReachProb[sameHandIndex]
		} else {
			removal = -cardRemoval[traverserHands[index].Hand[0]] - cardRemoval[traverserHands[index].Hand[1]]
		}
		utilities[index] = (probabilitySum + removal) * utility
	}
	return utilities
}

func (node *TerminalNode) PrintNodeDetails(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("TerminalNode last: %v pot %v oop %v ip %v\n",node.playerNode ^ 1, node.potSize, node.oopPlayerStack, node.ipPlayerStack)
}