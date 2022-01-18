// Package neatns contains Novelty Search implementation for NEAT method of ANN's evolving
package neatns

// how many nearest neighbors to consider for calculating novelty score?
const knnNoveltyScore = 15

// the maximal allowed size for fittest items list
const fittestAllowedSize = 5

// the minimal number of seed novelty items to start from
const archiveSeedAmount = 1

// NoveltyMetric The novelty metric function type.
// The function to compare two novelty items and return distance/difference between them
type NoveltyMetric func(x, y *NoveltyItem) float64

// NoveltyArchiveOptions defines options to be used by NoveltyArchive
type NoveltyArchiveOptions struct {
	// KNNNoveltyScore how many nearest neighbors to consider for calculating novelty score, i.e., for how many
	// neighbors to look at for N-nearest neighbor distance novelty
	KNNNoveltyScore int
	// FittestAllowedSize the maximal allowed size for fittest items list
	FittestAllowedSize int
	// ArchiveSeedAmount is the minimal number of seed novelty items to start from
	ArchiveSeedAmount int
}

// DefaultNoveltyArchiveOptions is to create default NoveltyArchiveOptions
func DefaultNoveltyArchiveOptions() NoveltyArchiveOptions {
	return NoveltyArchiveOptions{
		KNNNoveltyScore:    knnNoveltyScore,
		FittestAllowedSize: fittestAllowedSize,
		ArchiveSeedAmount:  archiveSeedAmount,
	}
}
