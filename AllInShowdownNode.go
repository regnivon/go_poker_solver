package solv

import (
	"fmt"
	//"os"
)

type AllInShowdownNode struct {
	potSize float64
	winUtility float64
	street int
	nextNodes []*ShowdownNode
	cache *RiverEvaluationCache
}

func NewAllInShowdownNode(potSize float64, street int,
						  cache *RiverEvaluationCache) *AllInShowdownNode {
	node := AllInShowdownNode{
		potSize: potSize,
		winUtility: potSize / 2.0,
		street: street,
		cache: cache,
	}
	node.nextNodes = make([]*ShowdownNode, 0)
	return &node
}

func (node *AllInShowdownNode) PrintNodeDetails(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("AllInShowdownNode street %v pot %v\n", node.street, node.potSize)
}

//TODO: check that it is unneeded to zero opp reach prob if overlap
func (node *AllInShowdownNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(traverserReachProb))
	for _, next := range node.nextNodes {
		newReach := make([]float64, len(opponentReachProb))
		for index, hand := range traversal.Ranges[traversal.Traverser ^ 1] {
			if !CheckHandBoardOverlap(hand.Hand, next.board) {
				newReach[index] = opponentReachProb[index]
			}
		}
		runoutEV := next.CFRTraversal(traversal, traverserReachProb, newReach)
		if node.street == 1 {
			for i := range runoutEV {
				utility[i] += runoutEV[i] * 2
			}
		} else {
			for i := range runoutEV {
				utility[i] += runoutEV[i]
			}
		}
	}
	if node.street == 1 {
		for i := range utility {
			utility[i] /= 1980.0
		}
	} else {
		for i := range utility {
			utility[i] /= 44.0
		}
	}
	return utility
}

func (node *AllInShowdownNode) AddNextNode(next *ShowdownNode) {
	node.nextNodes = append(node.nextNodes, next)
}

func (node *AllInShowdownNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(traversal.Ranges[traversal.Traverser]))
	for _, next := range node.nextNodes {
		newReach := make([]float64, len(opponentReachProb))
		for index, hand := range traversal.Ranges[traversal.Traverser ^ 1] {
			if !CheckHandBoardOverlap(hand.Hand, next.board) {
				newReach[index] = opponentReachProb[index]
			}
		}
		nodeUtility := next.BestResponse(traversal, newReach)
		if node.street == 1 {
			for i := range utility {
				utility[i] += nodeUtility[i] * 2
			}
		} else {
			for i := range utility {
				utility[i] += nodeUtility[i]
			}
		}
	}

	if node.street == 1 {
		for i := range utility {
			utility[i] /= 1980.0
		}
	} else {
		for i := range utility {
			utility[i] /= 44.0
		}
	}
	return utility
}