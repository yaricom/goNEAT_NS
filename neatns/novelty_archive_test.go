package neatns

import (
	"testing"
	"github.com/yaricom/goNEAT/neat/genetics"
	"strings"
)

const gnome_str = "genomestart 1\n" +
	"trait 1 0.1 0 0 0 0 0 0 0\n" +
	"trait 2 0.2 0 0 0 0 0 0 0\n" +
	"trait 3 0.3 0 0 0 0 0 0 0\n" +
	"node 1 0 1 1\n" +
	"node 2 0 1 1\n" +
	"node 3 0 1 3\n" +
	"node 4 0 0 2\n" +
	"gene 1 1 4 1.5 false 1 0 true\n" +
	"gene 2 2 4 2.5 false 2 0 true\n" +
	"gene 3 3 4 3.5 false 3 0 true\n" +
	"genomeend 1"

// tests archive update by fittest organisms
func TestNoveltyArchive_updateFittestWithOrganism(t *testing.T) {
	archive := NewNoveltyArchive(1.0, nil)

	// test normal update
	gen, err := genetics.ReadGenome(strings.NewReader(gnome_str), 1)
	if err != nil {
		t.Error(err)
	}
	org := genetics.NewOrganism(0.1, gen, 1)
	err = archive.updateFittestWithOrganism(fillOrganismData(org, 0.0))
	if err != nil {
		t.Error(err)
		return
	}

	if len(archive.FittestItems) != 1 {
		t.Errorf("len(archive.FittestItems) != 1, found: %d\n", len(archive.FittestItems))
		return
	}

	for i := 2; i <= fittestAllowedSize; i++ {
		archive.updateFittestWithOrganism(fillOrganismData(genetics.NewOrganism(0.1 * float64(i), gen, 1), 0.0))
	}

	for i := 0; i < fittestAllowedSize; i++ {
		if archive.FittestItems[i].Fitness != float64(fittestAllowedSize - i) * 0.1 {
			t.Errorf("Wrong item fitness: %f at index: %d\n", archive.FittestItems[i].Fitness, i)
		}
	}

	// test update over allowed size
	fitness := 0.6
	archive.updateFittestWithOrganism(fillOrganismData(genetics.NewOrganism(fitness, gen, 1), 0.0))
	if len(archive.FittestItems) != fittestAllowedSize {
		t.Error("len(archive.FittestItems) != fittestAllowedSize")
	}

	if archive.FittestItems[0].Fitness != fitness {
		t.Error("The item with maximal fitness at wrong position")
	}
}


func fillOrganismData(org *genetics.Organism, novelty float64) *genetics.Organism{
	ni := NoveltyItem{
		genome:org.Genotype,
		phenotype:org.Phenotype,
		Generation:org.Generation,
		Fitness:org.Fitness,
		Novelty:novelty,
	}
	org.Data = &genetics.OrganismData{Value:ni}
	return org
}
