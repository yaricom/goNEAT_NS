package neatns

import "container/list"

// The novelty archive contains all of the novel items we have encountered thus far.
// Using a novelty metric we can determine how novel a new item is compared to everything
// currently in the novelty set
type NoveltyArchive struct {
	// the all the novel items we have found so far
	NovelItems       []*NoveltyItem
	// the all novel items with fittest organisms associated found so far
	FittestItems     []*NoveltyItem

	// the current generation
	Generation int

	// the measure of novelty
	noveltyMetric    NoveltyMetric

	// the novel items waiting addition to the set pending the end of the generation
	itemsQueue       list.List

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

	}
	return &arch
}

// add novelty item to archive
func (a *NoveltyArchive) addNoveltyItem(i *NoveltyItem, addToQueue bool) {
	i.added = true
	i.generation = a.Generation
	a.NovelItems = append(a.NovelItems, i)
	if addToQueue {
		a.itemsQueue.PushBack(i)
	}
}
