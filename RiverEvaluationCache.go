package solv

import (
	"github.com/chehsunliu/poker"
	"sort"
)

type RiverEvaluationCache struct {
	ipRange Range
	oopRange Range
	indexCache map[[5]poker.Card]int
	RankingCache [][2][]HandRankPair
}

func NewRiverEvaluationCache(oopRange, ipRange Range) *RiverEvaluationCache{
	return &RiverEvaluationCache{
		ipRange: ipRange,
		oopRange: oopRange,
		indexCache: make(map[[5]poker.Card]int, 50),
		RankingCache: make([][2][]HandRankPair, 0, 50),
	}
}

func (cache *RiverEvaluationCache) InsertBoard(board []poker.Card) int {
	sort.Slice(board, func(i, j int) bool {
		return board[i] < board[j]
	})
	var boardCopy [5]poker.Card
	copy(boardCopy[:], board)
	if index, ok := cache.indexCache[boardCopy]; ok {
		return index
	}
	rankings := cache.FillHandRankings(board)
	cache.RankingCache = append(cache.RankingCache, rankings)
	cache.indexCache[boardCopy] = len(cache.RankingCache) - 1
	return len(cache.RankingCache) - 1
}

func (cache *RiverEvaluationCache) FillHandRankings(board []poker.Card) [2][]HandRankPair {
	ipRanks := cache.fillIPHandRankings(board)
	oopRanks := cache.fillOOPHandRankings(board)

	sort.Slice(ipRanks, func(i, j int) bool {
		return ipRanks[i].Rank > ipRanks[j].Rank
	})

	sort.Slice(oopRanks, func(i, j int) bool {
		return oopRanks[i].Rank > oopRanks[j].Rank
	})

	return [2][]HandRankPair{
		oopRanks,
		ipRanks,
	}
}

func (cache *RiverEvaluationCache) fillIPHandRankings(board []poker.Card) []HandRankPair {
	ipRanks := make([]HandRankPair, len(cache.ipRange))
	for i := range cache.ipRange {
		finalHand := append(board, cache.ipRange[i].Hand[0])
		finalHand = append(finalHand, cache.ipRange[i].Hand[1])
		ipRanks[i] = HandRankPair{cache.ipRange[i].Hand, poker.Evaluate(finalHand)}
	}
	return ipRanks
}

func (cache *RiverEvaluationCache) fillOOPHandRankings(board []poker.Card) []HandRankPair {
	oopRanks := make([]HandRankPair, len(cache.oopRange))
	for i := range cache.oopRange {
		finalHand := append(board, cache.oopRange[i].Hand[0])
		finalHand = append(finalHand, cache.oopRange[i].Hand[1])
		oopRanks[i] = HandRankPair{cache.oopRange[i].Hand, poker.Evaluate(finalHand)}
	}
	return oopRanks
}