package eval

// Category ordering matches proto: HIGH_CARD=1 ... STRAIGHT_FLUSH=9
// Higher category wins. If same category, compare RankVector lexicographically.
type Rank struct {
	Category   int32   // 1..9
	RankVector []int32 // tie-breakers, high-to-low
}

func Compare(a, b Rank) int {
	// returns: 1 if a>b, -1 if a<b, 0 if tie
	if a.Category > b.Category {
		return 1
	}
	if a.Category < b.Category {
		return -1
	}
	// same category: compare vectors lexicographically
	n := len(a.RankVector)
	if len(b.RankVector) < n {
		n = len(b.RankVector)
	}
	for i := 0; i < n; i++ {
		if a.RankVector[i] > b.RankVector[i] {
			return 1
		}
		if a.RankVector[i] < b.RankVector[i] {
			return -1
		}
	}
	// if all equal up to min length, longer vector wins (shouldn't matter if consistent)
	if len(a.RankVector) > len(b.RankVector) {
		return 1
	}
	if len(a.RankVector) < len(b.RankVector) {
		return -1
	}
	return 0
}
