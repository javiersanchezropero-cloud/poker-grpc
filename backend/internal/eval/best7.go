package eval

import "poker-grpc/backend/internal/cards"

// BestOf7 returns the best 5-card subset (as indices into the input slice)
// and its Rank. Input must have length 7.
func BestOf7(cs []cards.Card) (bestIdx [5]int, bestRank Rank) {
	// init with first combination [0,1,2,3,4]
	bestIdx = [5]int{0, 1, 2, 3, 4}
	bestRank = Eval5([]cards.Card{cs[0], cs[1], cs[2], cs[3], cs[4]})

	// generate all combinations of 5 out of 7 (C(7,5)=21)
	for a := 0; a < 3; a++ {
		for b := a + 1; b < 4; b++ {
			for c := b + 1; c < 5; c++ {
				for d := c + 1; d < 6; d++ {
					for e := d + 1; e < 7; e++ {
						idx := [5]int{a, b, c, d, e}
						r := Eval5([]cards.Card{cs[a], cs[b], cs[c], cs[d], cs[e]})
						if Compare(r, bestRank) == 1 {
							bestRank = r
							bestIdx = idx
						}
					}
				}
			}
		}
	}

	return bestIdx, bestRank
}
