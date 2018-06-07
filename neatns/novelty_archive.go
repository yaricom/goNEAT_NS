package neatns

import (
	"container/list"
	"github.com/yaricom/goNEAT/neat/genetics"
	"sort"
	"errors"
)

// The maximal allowed size for fittest items list
const fittestAllowedSize = 5

// The novelty archive contains all of the novel items we have encountered thus far.
// Using a novelty metric we can determine how novel a new item is compared to everything
// currently in the novelty set
type NoveltyArchive struct {
	// the all the novel items we have found so far
	NovelItems       []*NoveltyItem
	// the all novel items with fittest organisms associated found so far
	FittestItems     NoveltyItemsByFitness

	// the current generation
	Generation int

	// the measure of novelty
	noveltyMetric    NoveltyMetric

	// the novel items waiting addition to the set pending the end of the generation
	itemsQueue       *list.List

	// the minimum threshold for a "novel item"
	noveltyThreshold float64
	// the minimal possible value of novelty threshold
	noveltyFloor     float64

	// the counter to keep track of how many gens since we've added to the archive
	timeOut          int
	// the parameter for how many neighbors to look at for N-nearest neighbor distance novelty
	neighbors        int
}

// Creates new instance of novelty archive
func NewNoveltyArchive(threshold float64, metric NoveltyMetric) *NoveltyArchive {
	arch := NoveltyArchive{
		NovelItems:make([]*NoveltyItem, 0),
		FittestItems:make([]*NoveltyItem, 0),
		noveltyMetric:metric,
		itemsQueue:list.New(),
		neighbors:KNNNoveltyScore,
		noveltyFloor:0.25,
		noveltyThreshold:threshold,

	}
	return &arch
}

// add novelty item to archive
func (a *NoveltyArchive) addNoveltyItem(i *NoveltyItem, addToQueue bool) {
	i.added = true
	i.Generation = a.Generation
	a.NovelItems = append(a.NovelItems, i)
	if addToQueue {
		a.itemsQueue.PushBack(i)
	}
}

// to maintain list of fittest organisms so far
func (a *NoveltyArchive) updateFittestWithOrganism(org *genetics.Organism) error {
	if org.Data == nil {
		return errors.New("Organism with no Data provided")
	}

	if len(a.FittestItems) < fittestAllowedSize {
		// store organism's novelty item into fittest
		item := org.Data.Value.(NoveltyItem)
		a.FittestItems = append(a.FittestItems, &item)

		// sort to have most fit first
		sort.Sort(sort.Reverse(a.FittestItems))
	} else {
		last_item := a.FittestItems[len(a.FittestItems) - 1]
		org_item := org.Data.Value.(NoveltyItem)
		if org_item.Fitness > last_item.Fitness {
			// store organism's novelty item into fittest
			a.FittestItems = append(a.FittestItems, &org_item)

			// sort to have most fit first
			sort.Sort(sort.Reverse(a.FittestItems))

			// remove less fit item
			items := make([]*NoveltyItem, fittestAllowedSize)
			copy(items, a.FittestItems)
			a.FittestItems = items
		}
	}
	return nil
}
