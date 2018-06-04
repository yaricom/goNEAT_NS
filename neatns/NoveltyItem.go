// The package neatns contains Novelty Search implementation for NEAT method of ANN's evolving
package neatns

import (
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT/neat/network"
)

// The data holder for novel item's genome and phenotype
type NoveltyItem struct {
	// The genome of novel item
	genome    *genetics.Genome
	// The phenotype of novel item
	phenotype *network.Network
}
