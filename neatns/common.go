// Package neatns contains Novelty Search implementation for NEAT method of ANN's evolving
package neatns

// KNNNoveltyScore how many nearest neighbors to consider for calculating novelty score?
const KNNNoveltyScore = 15

// NoveltyMetric The novelty metric function type.
// The function to compare two novelty items and return distance/difference between them
type NoveltyMetric func(x, y *NoveltyItem) float64
