package neatns

import (
	"errors"
	"fmt"
	"github.com/yaricom/goNEAT/v2/neat"
	"github.com/yaricom/goNEAT/v2/neat/genetics"
	"io"
	"sort"
)

// The maximal allowed size for fittest items list
const fittestAllowedSize = 5

const archiveSeedAmount = 1

// NoveltyArchive The novelty archive contains all the novel items we have encountered thus far.
// Using a novelty metric we can determine how novel a new item is compared to everything
// currently in the novelty set
type NoveltyArchive struct {
	// the all the novel items we have found so far
	NovelItems []*NoveltyItem
	// the all novel items with the fittest organisms associated found so far
	FittestItems NoveltyItemsByFitness

	// the current generation
	Generation int

	// the measure of novelty
	noveltyMetric NoveltyMetric

	// the novel items added during current generation
	itemsAddedInGeneration int
	// the current generation index
	generationIndex int

	// the minimum threshold for a "novel item"
	noveltyThreshold float64
	// the minimal possible value of novelty threshold
	noveltyFloor float64

	// the counter to keep track of how many gens since we've added to the archive
	timeOut int

	// the parameter for how many neighbors to look at for N-nearest neighbor distance novelty
	neighbors int
}

// NewNoveltyArchive creates new instance of novelty archive
func NewNoveltyArchive(threshold float64, metric NoveltyMetric) *NoveltyArchive {
	arch := NoveltyArchive{
		NovelItems:       make([]*NoveltyItem, 0),
		FittestItems:     make([]*NoveltyItem, 0),
		noveltyMetric:    metric,
		neighbors:        KNNNoveltyScore,
		noveltyFloor:     0.25,
		noveltyThreshold: threshold,
		generationIndex:  archiveSeedAmount,
	}
	return &arch
}

// EvaluateIndividualNovelty evaluates the novelty of a single individual organism within population and update its fitness (onlyFitness = true)
// or store individual's novelty item into archive
func (a *NoveltyArchive) EvaluateIndividualNovelty(org *genetics.Organism, pop *genetics.Population, onlyFitness bool) {
	if org.Data == nil {
		neat.InfoLog(fmt.Sprintf(
			"WARNING! Found Organism without novelty point associated: %s\nNovelty evaluation will be skipped for it. Probably winner found!", org))
		return
	}
	item := org.Data.Value.(*NoveltyItem)
	var result float64
	if onlyFitness {
		// assign organism fitness according to average novelty within archive and population
		result = a.noveltyAvgKnn(item, -1, pop)
		org.Fitness = result
	} else {
		// consider adding a point to archive based on dist to nearest neighbor
		result = a.noveltyAvgKnn(item, 1, nil)
		if result > a.noveltyThreshold || len(a.NovelItems) < archiveSeedAmount {
			a.addNoveltyItem(item)
			item.Age += 1.0
		}
	}

	// store found values to the item
	item.Novelty = result
	item.Generation = a.Generation

	org.Data.Value = item
}

// EvaluatePopulationNovelty evaluates the novelty of the whole population and update organisms fitness (onlyFitness = true)
// or store each population individual's novelty items into archive
func (a *NoveltyArchive) EvaluatePopulationNovelty(pop *genetics.Population, onlyFitness bool) {
	for _, o := range pop.Organisms {
		a.EvaluateIndividualNovelty(o, pop, onlyFitness)
	}
}

// UpdateFittestWithOrganism to maintain list of the fittest organisms so far
func (a *NoveltyArchive) UpdateFittestWithOrganism(org *genetics.Organism) error {
	if org.Data == nil {
		return errors.New("organism with no Data provided")
	}

	if len(a.FittestItems) < fittestAllowedSize {
		// store organism's novelty item into fittest
		item := org.Data.Value.(*NoveltyItem)
		a.FittestItems = append(a.FittestItems, item)

		// sort to have most fit first
		sort.Sort(sort.Reverse(a.FittestItems))
	} else {
		lastItem := a.FittestItems[len(a.FittestItems)-1]
		orgItem := org.Data.Value.(*NoveltyItem)
		if orgItem.Fitness > lastItem.Fitness {
			// store organism's novelty item into fittest
			a.FittestItems = append(a.FittestItems, orgItem)

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

// EndOfGeneration the steady-state end of generation call
func (a *NoveltyArchive) EndOfGeneration() {
	a.Generation++

	a.adjustArchiveSettings()
}

// PrintNoveltyPoints prints collected novelty points to provided writer
func (a *NoveltyArchive) PrintNoveltyPoints(w io.Writer) error {
	if len(a.NovelItems) == 0 {
		return errors.New("no novel items to print")
	}
	for _, p := range a.NovelItems {
		str := p.String()
		if _, err := fmt.Fprintln(w, str); err != nil {
			return err
		}
	}
	return nil
}

// PrintFittest prints collected individuals with maximal fitness
func (a *NoveltyArchive) PrintFittest(w io.Writer) error {
	if len(a.FittestItems) == 0 {
		return errors.New("no fittest items to print")
	}
	for _, f := range a.FittestItems {
		str := f.String()
		if _, err := fmt.Fprintln(w, str); err != nil {
			return err
		}
	}
	return nil
}

// addNoveltyItem adds novelty item to archive
func (a *NoveltyArchive) addNoveltyItem(i *NoveltyItem) {
	i.added = true
	i.Generation = a.Generation
	a.NovelItems = append(a.NovelItems, i)
	a.itemsAddedInGeneration++
}

// adjustArchiveSettings is to adjust dynamic novelty threshold depending on how many have been added to archive recently
func (a *NoveltyArchive) adjustArchiveSettings() {
	if a.itemsAddedInGeneration == 0 {
		a.timeOut++
	} else {
		a.timeOut = 0
	}

	// if no individuals have been added for 10 generations lower the threshold
	if a.timeOut == 10 {
		a.noveltyThreshold *= 0.95
		if a.noveltyThreshold < a.noveltyFloor {
			a.noveltyThreshold = a.noveltyFloor
		}
		a.timeOut = 0
	}

	// if more than four individuals added this generation raise threshold
	if a.itemsAddedInGeneration >= 4 {
		a.noveltyThreshold *= 1.2
	}

	a.itemsAddedInGeneration = 0
	a.generationIndex = len(a.NovelItems)
}

// noveltyAvgKnn allows the K nearest neighbor novelty score calculation for given item within provided population
func (a *NoveltyArchive) noveltyAvgKnn(item *NoveltyItem, neigh int, pop *genetics.Population) float64 {
	var novelties ItemsDistances
	if pop != nil {
		novelties = a.mapNoveltyInPopulation(item, pop)
	} else {
		novelties = a.mapNovelty(item)
	}

	// sort by distance - minimal first
	sort.Sort(novelties)

	density, sum, weight := 0.0, 0.0, 0.0
	length := len(novelties)

	// if neighbors size not set - use value from archive parameters
	if neigh == -1 {
		neigh = a.neighbors
	}

	if length >= archiveSeedAmount {
		for i := 0; weight < float64(neigh) && i < len(novelties); i++ {
			sum += novelties[i].distance
			weight += 1.0
		}

		// find average
		if weight > 0 {
			density = sum / weight
		}
	}

	return density
}

// mapNovelty maps the novelty metric across the archive against provided item
func (a *NoveltyArchive) mapNovelty(item *NoveltyItem) ItemsDistances {
	distances := make([]ItemsDistance, len(a.NovelItems))
	for i := 0; i < len(a.NovelItems); i++ {
		distances[i] = ItemsDistance{
			distance: a.noveltyMetric(a.NovelItems[i], item),
			from:     a.NovelItems[i],
			to:       item,
		}
	}
	return distances
}

// mapNoveltyInPopulation maps the novelty metric across the archive and the current population
func (a *NoveltyArchive) mapNoveltyInPopulation(item *NoveltyItem, pop *genetics.Population) ItemsDistances {
	distances := make([]ItemsDistance, len(a.NovelItems))
	nIndex := 0
	for i := 0; i < len(a.NovelItems); i++ {
		distances[nIndex] = ItemsDistance{
			distance: a.noveltyMetric(a.NovelItems[i], item),
			from:     a.NovelItems[i],
			to:       item,
		}
		nIndex++
	}

	for i := 0; i < len(pop.Organisms); i++ {
		if pop.Organisms[i].Data != nil {
			orgItem := pop.Organisms[i].Data.Value.(*NoveltyItem)
			dist := ItemsDistance{
				distance: a.noveltyMetric(orgItem, item),
				from:     orgItem,
				to:       item,
			}
			distances = append(distances, dist)
		}
	}
	return distances
}
