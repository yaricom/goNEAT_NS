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

	// The flag to indicate whether item was added to archive
	added     bool
	// The generation when item was added to archive
	generation int
}
