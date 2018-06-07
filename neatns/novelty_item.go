package neatns

import (
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT/neat/network"
)

// The data holder for novel item's genome and phenotype
type NoveltyItem struct {
	// The genome of novel item
	genome     *genetics.Genome
	// The phenotype of novel item
	phenotype  *network.Network

	// The flag to indicate whether item was added to archive
	added      bool
	// The generation when item was added to archive
	Generation int

	// The fitness of the associated organism
	Fitness    float64
	// The novelty of this item
	Novelty    float64
}

// The sortable list of novelty items by fitness
type NoveltyItemsByFitness []*NoveltyItem
func (f NoveltyItemsByFitness) Len() int {
	return len(f)
}
func (f NoveltyItemsByFitness) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
func (f NoveltyItemsByFitness) Less(i, j int) bool {
	if f[i].Fitness < f[j].Fitness {
		return true
	} else if f[i].Fitness == f[j].Fitness {
		if f[i].Novelty < f[j].Novelty {
			return true // less novel is less
		}
	}

	return false
}
