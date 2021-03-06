package solv

import (
	"github.com/chehsunliu/poker"
	"regexp"
	"strconv"
	"strings"
)

const suits = "shdc"
const ranks = "23456789TJQKA"

func cardTo52Int(c poker.Card) int {
	rank := c.Rank()
	suit := c.Suit()
	if suit == 8 {
		suit = 3
	} else {
		suit = suit >> 1
	}
	return int( 4 * rank + suit)
}

func intToCard(i int) poker.Card {
	suit := i % 4
	rank := i / 4
	return poker.NewCard(string(ranks[rank]) + string(suits[suit]))
}

func constructPossibleNextCards(board []poker.Card, numNext int) []poker.Card {
	next := make([]poker.Card, numNext)
	count := 0
	for rankIndex := range ranks {
		for suitIndex := range suits {
			card := poker.NewCard(string(ranks[rankIndex]) + string(suits[suitIndex]))
			if !checkCardBoardOverlap(card, board) {
				next[count] = card
				count++
			}
		}
	}
	return next
}

func constructPossibleRunouts(board []poker.Card, cache *RiverEvaluationCache) [][]poker.Card {
	runouts := make([][]poker.Card, 0, 10)
	if len(board) == 3 {
		for i := 0; i < 52; i++ {
			card := intToCard(i)
			if checkCardBoardOverlap(card, board) {
				continue
			}
			for j := i+1; j < 52; j++ {
				riverCard := intToCard(j)
				if checkCardBoardOverlap(riverCard, board) {
					continue
				}
				boardCopy := make([]poker.Card, 5)
				copy(boardCopy, board)
				boardCopy[3] = card
				boardCopy[4] = riverCard
				runouts = append(runouts, boardCopy)
			}

		}
	} else if len(board) == 4 {
		for i := 0; i < 52; i++ {
			card := intToCard(i)
			if checkCardBoardOverlap(card, board) {
				continue
			}
			boardCopy := make([]poker.Card, 5)
			copy(boardCopy, board)
			boardCopy[4] = card
			runouts = append(runouts, boardCopy)
		}
	}
	return runouts
}

func convertRangeToFloatSlice(rng Range) []float64 {
	arr := make([]float64, len(rng))
	for index := range rng {
		arr[index] = rng[index].Combos
	}
	return arr
}

func checkCardBoardOverlap(card poker.Card, board []poker.Card) bool {
	for index := range board {
		if board[index] == card {
			return true
		}
	}
	return false
}

//CheckHandOverlap returns true if the hands overlap and false otherwise
func CheckHandOverlap(h1, h2 Hand) bool {
	return h1[0] == h2[0] || h1[1] == h2[0] || h1[0] == h2[1] || h1[1] == h2[1]
}

func CheckHandBoardOverlap(hand Hand, board []poker.Card) bool {
	for _, card := range hand {
		for _, boardCard := range board {
			if boardCard == card {
				return true
			}
		}
	}
	return false
}

func RemoveConflicts(handRange Range, board []poker.Card) Range {
	for i := 0;  i < len(handRange); i++ {
		if CheckHandBoardOverlap(handRange[i].Hand, board) {
			handRange = append(handRange[:i], handRange[i+1:]...)
			i--
		}
	}
	return handRange
}

//RangeRelativeProbabilities returns the probability of every hand in rng relative to the opponent cards
func RangeRelativeProbabilities(rng, oppRng Range) []float64 {
	normalizingValue := 0.0
	relatives := make([]float64, len(rng))

	for hand := range rng {
		prob := 0.0

		for otherHand := range oppRng {
			if !CheckHandOverlap(rng[hand].Hand, oppRng[otherHand].Hand) {
				prob += oppRng[otherHand].Combos
			}
		}
		relatives[hand] = prob * rng[hand].Combos
		normalizingValue += relatives[hand]
	}
	for i := range relatives {
		relatives[i] /= normalizingValue
	}
	return relatives
}

func UnblockedHands(rng, oop Range) []float64 {
	handCounts := make([]float64, len(rng))

	for hand := range rng {
		counts := 0.0
		for otherHand := range oop {
			if !CheckHandOverlap(rng[hand].Hand, oop[otherHand].Hand) {
				counts += oop[otherHand].Combos
			}
		}
		handCounts[hand] = counts
	}
	return handCounts
}

func HandsStringToHandRange(hands string) Range {
	percentageSplitter := regexp.MustCompile(`/\d+?\.\d+?/`)
	handRange := make(HandToFloatMap)
	percentages := percentageSplitter.FindAllString(hands, -1)
	percentages = append([]string{"/100.0/"}, percentages...)
	splits := percentageSplitter.Split(hands, -1)

	for index := range splits {
		processSplitAndPercentage(splits[index], percentages[index], handRange)
	}

	toReturn := make(Range, len(handRange))
	count := 0
	for hand := range handRange {
		handCombo := NewCombo(hand, handRange[hand])
		toReturn[count] = *handCombo
		count++
	}
	return toReturn
}

func processSplitAndPercentage(split, percentage string, handRange HandToFloatMap) {
	commaSplitter := regexp.MustCompile(`,\s`)
	percentageNumber, _ := strconv.ParseFloat(strings.Trim(percentage, "/"), 64)
	percentageNumber /= 100.0
	splits := commaSplitter.Split(split, -1)
	for index := range splits {
		current := splits[index]
		if len(current) > 0 {
			if strings.Contains(current, "+") {
				if current[0] == current[1] {
					processPairPlus(current, percentageNumber, handRange)
				} else {
					processHandPlus(current, percentageNumber, handRange)
				}
			} else if strings.Contains(current, "s") {
				processHandSuited(current, percentageNumber, handRange)
			} else if strings.Contains(current, "o") {
				processHandOffsuit(current, percentageNumber, handRange)
			} else if current[0] == current[1] {
				processPair(current, percentageNumber, handRange)
			} else {
				processHandBoth(current, percentageNumber, handRange)
			}
		}
	}
}

func processHandPlus(base string, percentage float64, handRange HandToFloatMap) {
	ranks := "23456789TJQKA"
	suited := strings.Contains(base, "s")
	offsuit := strings.Contains(base, "o")
	startIndex := strings.Index(ranks, string(base[1]))
	for i := startIndex; i < len(ranks); i++ {
		currentBase := string(base[0]) + string(ranks[i])
		if !suited && !offsuit {
			processHandBoth(currentBase, percentage, handRange)
		} else if suited {
			processHandSuited(currentBase, percentage, handRange)
		} else {
			processHandOffsuit(currentBase, percentage, handRange)
		}
	}
}

func processHandSuited(base string, percentage float64, handRange HandToFloatMap) {
	c1 := string(base[0])
	c2 := string(base[1])
	for index := range suits {
		hand := NewHand(c1 + string(suits[index]), c2 + string(suits[index]))
		handRange[hand] = percentage
	}
}

func processHandOffsuit(base string, percentage float64, handRange HandToFloatMap) {
	c1 := string(base[0])
	c2 := string(base[1])
	for firstIndex := range suits {
		for secondIndex := range suits {
			if firstIndex != secondIndex {
				hand := NewHand(c1 + string(suits[firstIndex]), c2 + string(suits[secondIndex]))
				handRange[hand] = percentage
			}
		}
	}
}

func processHandBoth(base string, percentage float64, handRange HandToFloatMap) {
	processHandSuited(base, percentage, handRange)
	processHandOffsuit(base, percentage, handRange)
}

func processPair(base string, percentage float64, handRange HandToFloatMap) {
	pairCard := string(base[0])

	handRange[NewHand(pairCard + "c", pairCard + "d")] = percentage
	handRange[NewHand(pairCard + "c", pairCard + "s")] = percentage
	handRange[NewHand(pairCard + "c", pairCard + "h")] = percentage
	handRange[NewHand(pairCard + "h", pairCard + "d")] = percentage
	handRange[NewHand(pairCard + "h", pairCard + "s")] = percentage
	handRange[NewHand(pairCard + "d", pairCard + "s")] = percentage
}

func processPairPlus(base string, percentage float64, handRange HandToFloatMap) {
	ranks := "23456789TJQKA"
	startIndex := strings.Index(ranks, string(base[0]))
	for i := startIndex; i < len(ranks); i++ {
		currentBase := string(ranks[i]) + string(ranks[i])
		processPair(currentBase, percentage, handRange)
	}
}

