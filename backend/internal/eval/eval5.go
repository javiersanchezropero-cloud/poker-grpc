package eval

import (
	"sort"

	"poker-grpc/backend/internal/cards"
)

// Eval5 evaluates exactly 5 cards and returns a comparable Rank.
// It assumes the cards are valid (no duplicates, valid ranks/suits).
func Eval5(cs []cards.Card) Rank {
	// counts by rank
	rankCount := map[int]int{}
	suitCount := map[string]int{}
	ranks := make([]int, 0, 5)

	for _, c := range cs {
		rankCount[c.Rank]++
		suitCount[c.Suit]++
		ranks = append(ranks, c.Rank)
	}

	// sort ranks desc
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] > ranks[j] })

	isFlush := false
	for _, cnt := range suitCount {
		if cnt == 5 {
			isFlush = true
			break
		}
	}

	isStraight, straightHigh := detectStraight(ranks)

	// Straight flush
	if isFlush && isStraight {
		return Rank{Category: 9, RankVector: []int32{int32(straightHigh)}}
	}

	// Group ranks by counts: prepare (count, rank) pairs
	type grp struct {
		cnt  int
		rank int
	}
	groups := make([]grp, 0, len(rankCount))
	for r, c := range rankCount {
		groups = append(groups, grp{cnt: c, rank: r})
	}
	// sort by: cnt desc, then rank desc
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].cnt != groups[j].cnt {
			return groups[i].cnt > groups[j].cnt
		}
		return groups[i].rank > groups[j].rank
	})

	// Four of a kind
	if groups[0].cnt == 4 {
		quad := groups[0].rank
		kicker := groups[1].rank
		return Rank{Category: 8, RankVector: []int32{int32(quad), int32(kicker)}}
	}

	// Full house
	if groups[0].cnt == 3 && groups[1].cnt == 2 {
		trips := groups[0].rank
		pair := groups[1].rank
		return Rank{Category: 7, RankVector: []int32{int32(trips), int32(pair)}}
	}

	// Flush
	if isFlush {
		return Rank{Category: 6, RankVector: toI32(ranks)}
	}

	// Straight
	if isStraight {
		return Rank{Category: 5, RankVector: []int32{int32(straightHigh)}}
	}

	// Three of a kind
	if groups[0].cnt == 3 {
		trips := groups[0].rank
		kickers := make([]int, 0, 2)
		for _, g := range groups[1:] {
			for i := 0; i < g.cnt; i++ {
				kickers = append(kickers, g.rank)
			}
		}
		sort.Slice(kickers, func(i, j int) bool { return kickers[i] > kickers[j] })
		return Rank{Category: 4, RankVector: []int32{int32(trips), int32(kickers[0]), int32(kickers[1])}}
	}

	// Two pair
	if groups[0].cnt == 2 && groups[1].cnt == 2 {
		highPair := groups[0].rank
		lowPair := groups[1].rank
		kicker := groups[2].rank
		return Rank{Category: 3, RankVector: []int32{int32(highPair), int32(lowPair), int32(kicker)}}
	}

	// One pair
	if groups[0].cnt == 2 {
		pair := groups[0].rank
		kickers := make([]int, 0, 3)
		for _, g := range groups[1:] {
			for i := 0; i < g.cnt; i++ {
				kickers = append(kickers, g.rank)
			}
		}
		sort.Slice(kickers, func(i, j int) bool { return kickers[i] > kickers[j] })
		return Rank{Category: 2, RankVector: []int32{int32(pair), int32(kickers[0]), int32(kickers[1]), int32(kickers[2])}}
	}

	// High card
	return Rank{Category: 1, RankVector: toI32(ranks)}
}

// detectStraight expects ranks sorted DESC but may include duplicates.
// Returns (isStraight, highCardOfStraight).
// Handles wheel straight A-2-3-4-5 as high=5.
func detectStraight(ranksDesc []int) (bool, int) {
	// unique ranks in desc order
	uniq := make([]int, 0, 5)
	seen := map[int]bool{}
	for _, r := range ranksDesc {
		if !seen[r] {
			seen[r] = true
			uniq = append(uniq, r)
		}
	}
	if len(uniq) != 5 {
		return false, 0
	}

	// normal straight: r, r-1, r-2, r-3, r-4
	if uniq[0]-uniq[4] == 4 {
		ok := true
		for i := 0; i < 4; i++ {
			if uniq[i]-uniq[i+1] != 1 {
				ok = false
				break
			}
		}
		if ok {
			return true, uniq[0]
		}
	}

	// wheel: A,5,4,3,2 (i.e. ranks are [14,5,4,3,2])
	if uniq[0] == 14 && uniq[1] == 5 && uniq[2] == 4 && uniq[3] == 3 && uniq[4] == 2 {
		return true, 5
	}

	return false, 0
}

func toI32(rs []int) []int32 {
	out := make([]int32, 0, len(rs))
	for _, r := range rs {
		out = append(out, int32(r))
	}
	return out
}
