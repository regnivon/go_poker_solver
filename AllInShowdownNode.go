package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
)

type AllInShowdownNode struct {
	potSize float64
	winUtility float64
	street int
	runoutIndices []int
	runouts [][]poker.Card
	cache *RiverEvaluationCache
}

func NewAllInShowdownNode(potSize float64, street int, board []poker.Card,
						  cache *RiverEvaluationCache) *AllInShowdownNode {
	runouts, indices := constructPossibleRunouts(board, cache)
	node := AllInShowdownNode{
		potSize: potSize,
		winUtility: potSize / 2.0,
		street: street,
		runoutIndices: indices,
		runouts: runouts,
		cache: cache,
	}
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
	for index := range node.runouts {
		traverserRanks := node.cache.RankingCache[index][traversal.Traverser]
		opponentRanks := node.cache.RankingCache[index][traversal.Traverser ^ 1]
		node.winnerShowdownProbabilityCalculation(traversal, utility, opponentReachProb, traverserRanks, opponentRanks)
		node.loserShowdownProbabilityCalculation(traversal, utility, opponentReachProb, traverserRanks, opponentRanks)
	}
	if node.street == 1 {
		for i := range utility {
			utility[i] /= 45.0
		}
	} else {
		for i := range utility {
			utility[i] /= 44.0
		}
	}
	return utility
}

func (node *AllInShowdownNode) winnerShowdownProbabilityCalculation(traversal *Traversal, utility, OpponentReachProb []float64,
																	TraverserRanks, OpponentRanks []HandRankPair) {
	cardRemoval := make(map[poker.Card]float64)
	winnerProbabilitySum := 0.0

	opponent := traversal.Traverser ^ 1

	opIndex := 0

	//iterate through all of the traverser's hands, and then while we have a better hand (lower rank) increase
	//the winProbability and account for hole card clashes
	for traverserRankIndex := range TraverserRanks {
		for opIndex < len(OpponentRanks) && OpponentRanks[opIndex].Rank > TraverserRanks[traverserRankIndex].Rank {
			prob := OpponentReachProb[traversal.IndexCaches[opponent][OpponentRanks[opIndex].Hand]]
			winnerProbabilitySum += prob
			cardRemoval[OpponentRanks[opIndex].Hand[0]] += prob
			cardRemoval[OpponentRanks[opIndex].Hand[1]] += prob

			opIndex++
		}
		utility[traversal.IndexCaches[traversal.Traverser][TraverserRanks[traverserRankIndex].Hand]] +=
			(winnerProbabilitySum -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[0]] -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[1]]) * node.winUtility
	}
}

func (node *AllInShowdownNode) loserShowdownProbabilityCalculation(traversal *Traversal, utility, OpponentReachProb []float64,
																   TraverserRanks, OpponentRanks []HandRankPair) {
	cardRemoval := make(map[poker.Card]float64)
	loserProbabilitySum := 0.0

	opponent := traversal.Traverser ^ 1

	opIndex := len(OpponentReachProb) - 1

	//iterate through all of the traverser's hands, and then while we have a better hand (lower rank) increase
	//the winProbability and account for hole card clashes
	for traverserRankIndex := len(TraverserRanks) - 1; traverserRankIndex >= 0 ; traverserRankIndex-- {
		for opIndex >= 0 && OpponentRanks[opIndex].Rank < TraverserRanks[traverserRankIndex].Rank {
			prob := OpponentReachProb[traversal.IndexCaches[opponent][OpponentRanks[opIndex].Hand]]

			loserProbabilitySum += prob
			cardRemoval[OpponentRanks[opIndex].Hand[0]] += prob
			cardRemoval[OpponentRanks[opIndex].Hand[1]] += prob

			opIndex--
		}
		utility[traversal.IndexCaches[traversal.Traverser][TraverserRanks[traverserRankIndex].Hand]] -=
			(loserProbabilitySum -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[0]] -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[1]]) * node.winUtility
	}
}

func (node *AllInShowdownNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(traversal.Ranges[traversal.Traverser]))
	for index := range node.runouts {
		traverserRanks := node.cache.RankingCache[index][traversal.Traverser]
		opponentRanks := node.cache.RankingCache[index][traversal.Traverser ^ 1]
		node.winnerShowdownProbabilityCalculation(traversal, utility, opponentReachProb, traverserRanks, opponentRanks)
		node.loserShowdownProbabilityCalculation(traversal, utility, opponentReachProb, traverserRanks, opponentRanks)
	}
	if node.street == 1 {
		for i := range utility {
			utility[i] /= 45.0
		}
	} else {
		for i := range utility {
			utility[i] /= 44.0
		}
	}
	return utility
}