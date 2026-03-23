package main

import "sync"

var rowPool = sync.Pool{
	New: func() interface{} {
		// A slice with capacity for the max expected word length
		return make([]int, 0, 100)
	},
}

// EditDistance finds the edit distance between two words using limited space.
// It also takes a maxDist for early exit optimization.
func EditDistance(s1, s2 string, maxDist int) int {
	len1 := len(s1)
	len2 := len(s2)

	// If length difference > maxDist, it's impossible to stay within limit
	if abs(len1-len2) > maxDist {
		return maxDist + 1
	}

	// We only need two rows instead of a full DP table
	row1 := rowPool.Get().([]int)[:0]
	row2 := rowPool.Get().([]int)[:0]

	// Ensure they have enough capacity
	if cap(row1) < len2+1 {
		row1 = make([]int, 0, len2+1)
	}
	if cap(row2) < len2+1 {
		row2 = make([]int, 0, len2+1)
	}

	// Re-slice to current capacity
	row1 = row1[:len2+1]
	row2 = row2[:len2+1]

	defer func() {
		rowPool.Put(row1)
		rowPool.Put(row2)
	}()

	// Initialize first row
	for j := 0; j <= len2; j++ {
		row1[j] = j
	}

	for i := 1; i <= len1; i++ {
		row2[0] = i
		minRowVal := i
		for j := 1; j <= len2; j++ {
			if s1[i-1] == s2[j-1] {
				row2[j] = row1[j-1]
			} else {
				row2[j] = 1 + min(row1[j], row2[j-1], row1[j-1])
			}
			if row2[j] < minRowVal {
				minRowVal = row2[j]
			}
		}
		// Early exit pruning: if every value in the current row > maxDist,
		// then further edits will only increase or maintain the distance.
		if minRowVal > maxDist {
			return maxDist + 1
		}
		// Copy row2 to row1 for next iteration
		copy(row1, row2)
	}

	return row1[len2]
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
