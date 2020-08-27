package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
)

//HandRankPair pairs a hand with its Rank, used for the O(N) showdown evaluation using a sorted list of
//ranks
type HandRankPair struct {
	Hand Hand
	Rank int32
}

//ShowdownNode is a GameNode that occurs in the game tree after the prior node action is either the IP player
//checking back the river, or after a call action. Thus, this node will calculate EVs for hands by
//comparing vs the opponents set of hands
type ShowdownNode struct {
	lastPlayer int
	winUtility float64
	cacheIndex int
	cache *RiverEvaluationCache
	board []poker.Card
}

//NewShowdownNode constructs a ShowdownNode
func NewShowdownNode(potSize float64, lastPlayer int, board []poker.Card, cache *RiverEvaluationCache) *ShowdownNode {
	node := ShowdownNode{
		lastPlayer: lastPlayer,
		cache: cache,
	}
	node.winUtility = potSize / 2.0
	node.board = board
	node.cacheIndex = 0
	return &node
}

func (node *ShowdownNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	return node.GetUtil(traversal, opponentReachProb)
}

//GetUtil calculates the utility for the traverser with the efficient algorithm
func (node *ShowdownNode) GetUtil(traversal *Traversal, opponentReachProb []float64) []float64 {
	if traversal.Traverser == 1 {
		return node.Showdown(traversal, node.cache.RankingCache[node.cacheIndex][1],
			node.cache.RankingCache[node.cacheIndex][0], opponentReachProb)
	}
	return node.Showdown(traversal, node.cache.RankingCache[node.cacheIndex][0],
		node.cache.RankingCache[node.cacheIndex][1], opponentReachProb)
}

func (node *ShowdownNode) PrintNodeDetails(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("Showdown node: last: %v pot %v\n", node.lastPlayer, node.winUtility * 2.0)
}

//Showdown calculates the hand utility for the traverser using the O(n) evaluation algorithm
func (node *ShowdownNode) Showdown(traversal *Traversal, TraverserRanks,
	 							   OpponentRanks []HandRankPair, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(traversal.Ranges[traversal.Traverser]))
	node.winnerShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	node.loserShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	return utility
}

func (node *ShowdownNode) winnerShowdownProbabilityCalculation(traversal *Traversal, utility, OpponentReachProb []float64,
															   TraverserRanks, OpponentRanks []HandRankPair) {
	cardRemoval := make(map[poker.Card]float64)
	winnerProbabilitySum := 0.0

	opponent := traversal.Traverser ^ 1

	opIndex := 0

	//iterate through all of the traverser's hands, and then while we have a better hand (lower rank) increase
	//the winProbability and account for hole card clashes
	for traverserRankIndex := range TraverserRanks {
		/*TODO: This section is a complete disaster since this calculation needs to happen quite often, so I am not
		sure if assigning all these variables nicely will come with a cost, will need to check this
		probably low hanging fruit
		*/
		for opIndex < len(OpponentRanks) && OpponentRanks[opIndex].Rank > TraverserRanks[traverserRankIndex].Rank {
			prob := OpponentReachProb[traversal.IndexCaches[opponent][OpponentRanks[opIndex].Hand]]
			winnerProbabilitySum += prob
			cardRemoval[OpponentRanks[opIndex].Hand[0]] += prob
			cardRemoval[OpponentRanks[opIndex].Hand[1]] += prob

			opIndex++
		}
		utility[traversal.IndexCaches[traversal.Traverser][TraverserRanks[traverserRankIndex].Hand]] =
			(winnerProbabilitySum -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[0]] -
				cardRemoval[TraverserRanks[traverserRankIndex].Hand[1]]) * node.winUtility
	}
}

func (node *ShowdownNode) loserShowdownProbabilityCalculation(traversal *Traversal, utility, OpponentReachProb []float64,
														      TraverserRanks, OpponentRanks []HandRankPair) {
	cardRemoval := make(map[poker.Card]float64)
	loserProbabilitySum := 0.0

	opponent := traversal.Traverser ^ 1

	opIndex := len(OpponentRanks) - 1

	//iterate through all of the traverser's hands, and then while we have a better hand (lower rank) increase
	//the winProbability and account for hole card clashes
	for traverserRankIndex := len(TraverserRanks) - 1; traverserRankIndex >= 0 ; traverserRankIndex-- {
		/*TODO: This section is a complete disaster since this calculation needs to happen quite often, so I am not
		sure if assigning all these variables nicely will come with a cost, will need to check this
		probably low hanging fruit
		*/
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

func (node *ShowdownNode) ShowdownSlow(traversal *Traversal, TraverserRanks,
									   OpponentRanks []HandRankPair, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(TraverserRanks))

	op := traversal.Traverser ^ 1
	trav := traversal.Traverser

	for traverserIndex := range TraverserRanks {
		probabilitySum := 0.0
		for oppIndex := range OpponentRanks {
			if !CheckHandOverlap(OpponentRanks[oppIndex].Hand, TraverserRanks[traverserIndex].Hand) {
				if OpponentRanks[oppIndex].Rank > TraverserRanks[traverserIndex].Rank {
					probabilitySum += opponentReachProb[traversal.IndexCaches[op][OpponentRanks[oppIndex].Hand]]
				} else if OpponentRanks[oppIndex].Rank < TraverserRanks[traverserIndex].Rank {
					probabilitySum -= opponentReachProb[traversal.IndexCaches[op][OpponentRanks[oppIndex].Hand]]
				}
			}
		}
		utility[traversal.IndexCaches[trav][TraverserRanks[traverserIndex].Hand]] = probabilitySum * node.winUtility
	}
	return utility
}

func (node *ShowdownNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	var TraverserRanks []HandRankPair
	var OpponentRanks []HandRankPair
	if traversal.Traverser == 0 {
		TraverserRanks = node.cache.RankingCache[node.cacheIndex][0]
		OpponentRanks = node.cache.RankingCache[node.cacheIndex][1]
	} else {
		OpponentRanks = node.cache.RankingCache[node.cacheIndex][0]
		TraverserRanks = node.cache.RankingCache[node.cacheIndex][1]
	}
	utility := make([]float64, len(traversal.Ranges[traversal.Traverser]))
	node.winnerShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	node.loserShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	return utility
}