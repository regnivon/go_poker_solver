package solv

import (
	"github.com/chehsunliu/poker"
	"sync"
)

//ChanceNode - important notes are the fact that nextCards[i] tells you which card nextNodes[i] represents
type ChanceNode struct {
	*GameNode
	nextCards []poker.Card
}

func NewChanceNode(gn *GameNode, board []poker.Card, street int) *ChanceNode {
	if street == 1 {
		return &ChanceNode{
			GameNode:  gn,
			nextCards: constructPossibleNextCards(board, 49),
		}
	} else {
		return &ChanceNode{
			GameNode:  gn,
			nextCards: constructPossibleNextCards(board, 48),
		}
	}
}

func (node *ChanceNode) CFRTraversal(traversal *Traversal, traverserReachProb, opponentReachProb []float64) []float64 {
	subResults := make([][]float64, len(node.nextCards))
	result := make([]float64, len(traverserReachProb))
	oppHands := traversal.Ranges[traversal.Traverser ^ 1]

	weights := node.GetTraverserHandWeightingForCard(traversal, opponentReachProb)

	var wg sync.WaitGroup
	wg.Add(len(node.nextNodes))
	for index, next := range node.nextNodes {
		go func(i int, nextNode Node) {
			defer wg.Done()
			nextTrav := make([]float64, len(traverserReachProb))
			nextOpp := make([]float64, len(opponentReachProb))
			for hand := range nextTrav {
				nextTrav[hand] = traverserReachProb[hand] * weights[hand + i * len(traverserReachProb)]
			}
			for hand := range nextOpp {
				if !(oppHands[hand].Hand[0] == node.nextCards[i] || oppHands[hand].Hand[1] == node.nextCards[i]) {
					nextOpp[hand] = opponentReachProb[hand]
				}
			}
			subResults[i] = nextNode.CFRTraversal(traversal, nextTrav, nextOpp)
		}(index, next)
	}
	wg.Wait()

	for index := range subResults {
		for hand := range result {
			result[hand] += subResults[index][hand]
		}
	}
	//TODO: this needs to be variable based on the street it should be 45 for flop
	//This is because board has 4 cards, we have 2, opp has 2, thus 52-4-2-2 = 44 is the number of
	//actual possible cards in a given hand
	for hand := range result {
		result[hand] /= 44
	}

	return result
}


//GetTraverserHandWeightingForCard - for each of the traverser's hands, calculate the
//weight of each river card given the opponents range. Basically, we need to take into account the chance that
//a card will come out while holding a given hand, respecting the fact that if we are holding a particular card
//we might be blocking our opponent from blocking a turn/river card. This will be utilized
//to re-weight the traverser reach probabilities for each possible turn/river card.
func (node *ChanceNode) GetTraverserHandWeightingForCard(traversal *Traversal, opponentReachProb []float64) []float64  {
	travHands := traversal.Ranges[traversal.Traverser]
	oppHands := traversal.Ranges[traversal.Traverser ^ 1]
	weight := make([]float64, len(node.nextCards) * len(travHands))

	cardRemoval := make([]float64, 52)
	probabilitySum := 0.0

	for handIndex := range opponentReachProb {
		cardRemoval[cardTo52Int(oppHands[handIndex].Hand[0])] += opponentReachProb[handIndex]
		cardRemoval[cardTo52Int(oppHands[handIndex].Hand[1])] += opponentReachProb[handIndex]

		probabilitySum += opponentReachProb[handIndex]
	}


	for handIndex := range travHands {
		weightSum := 0.0
		for card := range node.nextCards {
			hand := travHands[handIndex].Hand
			if hand[0] == node.nextCards[card] || hand[1] == node.nextCards[card] {
				continue
			}
			var curWeight float64
			sameHandIndex, ok := traversal.IndexCaches[traversal.Traverser ^ 1][hand]
			if ok {
				curWeight = probabilitySum -
					cardRemoval[cardTo52Int(node.nextCards[card])] -
					cardRemoval[cardTo52Int(hand[0])] -
					cardRemoval[cardTo52Int(hand[1])] + opponentReachProb[sameHandIndex]
			} else {
				curWeight = probabilitySum -
					cardRemoval[cardTo52Int(node.nextCards[card])] -
					cardRemoval[cardTo52Int(hand[0])] -
					cardRemoval[cardTo52Int(hand[1])]
			}
			weight[handIndex + card * len(travHands)] = curWeight
			weightSum += curWeight
		}

		for card := range node.nextCards {
			if weightSum > 0 {
				weight[handIndex + card * len(travHands)] /= weightSum
			}
		}
	}
	return weight
}
