package solv

import (
	"fmt"
	"github.com/chehsunliu/poker"
	"sort"
)

//TODO: init the hand rank pairs somewhere

//HandRankPair pairs a hand with its Rank, used for the O(N) showdown evaluation using a shorted list of
//ranks
type HandRankPair struct {
	Hand Hand
	Rank int32
}

//ShowdownNode is a GameNode that occurs in the game tree after the prior node action is either the IP player
//checking back the river, or after a call action. Thus, this node will calculate EVs for hands by
//comparing vs the opponents set of hands
type ShowdownNode struct {
	*GameNode
	winUtility float64
	ipRanks []HandRankPair
	oopRanks []HandRankPair
	board []poker.Card
}

//NewShowdownNode constructs a ShowdownNode
func NewShowdownNode(gameNode *GameNode) *ShowdownNode {
	node := ShowdownNode{GameNode: gameNode}
	node.isTerminal = true
	node.winUtility = node.potSize / 2.0
	node.board = []poker.Card{
		poker.NewCard("Ac"),
		poker.NewCard("7s"),
		poker.NewCard("5s"),
		poker.NewCard("3d"),
		poker.NewCard("2h"),
	}
	return &node
}

func (node *ShowdownNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	return node.GetUtil(traversal, traverserReachProb, opponentReachProb)
}

//GetUtil calculates the utility for the traverser with the efficient algorithm
func (node *ShowdownNode) GetUtil(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	if traversal.Traverser == 1 {
		return node.Showdown(traversal, node.ipRanks, node.oopRanks, opponentReachProb)
	}
	return node.Showdown(traversal, node.oopRanks, node.ipRanks, opponentReachProb)
}

//Showdown calculates the hand utility for the traverser using the O(n) evaluation algorithm
func (node *ShowdownNode) Showdown(traversal *Traversal, TraverserRanks,
	 							   OpponentRanks []HandRankPair, opponentReachProb []float64) []float64 {
	utility := make([]float64, len(TraverserRanks))
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

	opIndex := len(OpponentReachProb) - 1

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

//FillHandRankings is called upon the first reach of this showdown node, it evaluates all hands in each
//player's range then sorts the ranks for the O(n) utility calculation. Note that lower rank is better
func (node *ShowdownNode) FillHandRankings(IPHands, OOPHands Range) {
	node.ipRanks = make([]HandRankPair, len(IPHands))
	node.oopRanks = make([]HandRankPair, len(OOPHands))

	node.fillIPHandRankings(IPHands)
	node.fillOOPHandRankings(OOPHands)

	sort.Slice(node.ipRanks, func(i, j int) bool {
		return node.ipRanks[i].Rank > node.ipRanks[j].Rank
	})

	sort.Slice(node.oopRanks, func(i, j int) bool {
		return node.oopRanks[i].Rank > node.oopRanks[j].Rank
	})
}

func (node *ShowdownNode) fillIPHandRankings(IPHands Range) {
	for i := range IPHands {
		finalHand := append(node.board, IPHands[i].Hand[0])
		finalHand = append(finalHand, IPHands[i].Hand[1])
		node.ipRanks[i] = HandRankPair{IPHands[i].Hand, poker.Evaluate(finalHand)}
	}
}

func (node *ShowdownNode) fillOOPHandRankings(OOPHands Range) {
	for i := range OOPHands {
		finalHand := append(node.board, OOPHands[i].Hand[0])
		finalHand = append(finalHand, OOPHands[i].Hand[1])
		node.oopRanks[i] = HandRankPair{OOPHands[i].Hand, poker.Evaluate(finalHand)}
	}
}

//TODO: this should use cached ranks
func (node *ShowdownNode) BestResponse(traversal *Traversal, opponentReachProb []float64) []float64 {
	var TraverserRanks []HandRankPair
	var OpponentRanks []HandRankPair
	if traversal.Traverser == 0 {
		TraverserRanks = node.oopRanks
		OpponentRanks = node.ipRanks
	} else {
		OpponentRanks= node.oopRanks
		TraverserRanks= node.ipRanks
	}
	utility := make([]float64, len(TraverserRanks))
	node.winnerShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	node.loserShowdownProbabilityCalculation(traversal, utility, opponentReachProb, TraverserRanks, OpponentRanks)
	return utility
}

func (node *ShowdownNode) PrintNodeDetails() {
	fmt.Printf("Showdown node: last: %v pot %v oop %v ip %v\n",node.playerNode ^ 1, node.potSize, node.oopPlayerStack, node.ipPlayerStack)
}