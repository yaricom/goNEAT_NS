// The experiments package holds various experiments with NEAT Novelty Search.
package experiments

import "math"

// calculates item-wise difference between two vectors
func HistDiff(in1, in2 []float64) float64 {
	size := len(in1)
	diff_accum := 0.0
	for i := 0; i < size; i++ {
		diff := in1[i] - in2[i]
		diff_accum += math.Abs(diff)
	}
	return diff_accum / float64(size)
}
