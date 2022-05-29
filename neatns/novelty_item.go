package neatns

import (
	"fmt"
)

// NoveltyItem is the data holder for novel item's genome and phenotype
type NoveltyItem struct {
	// The flag to indicate whether item was added to archive
	added bool
	// The generation when item was added to archive
	Generation int `json:"generation"`

	// The ID of associated individual organism */
	IndividualID int `json:"individual_id"`

	// The fitness of the associated organism
	Fitness float64 `json:"fitness"`
	// The novelty of this item
	Novelty float64 `json:"novelty"`
	// The item's age
	Age float64 `json:"age"`

	// The data associated with item
	Data []float64 `json:"data"`
}

// NewNoveltyItem creates new novelty item
func NewNoveltyItem() *NoveltyItem {
	return &NoveltyItem{Data: make([]float64, 0)}
}

// Stringer
func (ni NoveltyItem) String() string {
	str := fmt.Sprintf("Novelty: %.2f Fitness: %f Generation: %d Individual: %d\n",
		ni.Novelty, ni.Fitness, ni.Generation, ni.IndividualID)
	str += "\tPoint: "
	for _, v := range ni.Data {
		str += fmt.Sprintf(" %.3f", v)
	}
	return str
}

// ItemsDistance the structure to hold distance between two items
type ItemsDistance struct {
	distance float64
	from, to *NoveltyItem
}

// ItemsDistances the sortable list of distances between two items
type ItemsDistances []ItemsDistance

func (f ItemsDistances) Len() int {
	return len(f)
}
func (f ItemsDistances) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
func (f ItemsDistances) Less(i, j int) bool {
	return f[i].distance < f[j].distance
}

// NoveltyItemsByFitness the sortable list of novelty items by fitness
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
