package neatns

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/neat/genetics"
	"math/rand"
	"strings"
	"testing"
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
	require.NoError(t, err, "failed to read genome")

	org, err := genetics.NewOrganism(0.1, gen, 1)
	require.NoError(t, err, "failed to create new organism")

	err = archive.UpdateFittestWithOrganism(fillOrganismData(org, 0.0))
	require.NoError(t, err, "failed to update fittest")
	require.Len(t, archive.FittestItems, 1)

	for i := 2; i <= fittestAllowedSize; i++ {
		org, err = genetics.NewOrganism(0.1*float64(i), gen, 1)
		require.NoError(t, err, "failed to create new organism at: %d", i)
		err = archive.UpdateFittestWithOrganism(fillOrganismData(org, 0.0))
		require.NoError(t, err, "failed to update fittest at: %d", i)
	}

	for i := 0; i < fittestAllowedSize; i++ {
		expected := float64(fittestAllowedSize-i) * 0.1
		assert.Equal(t, expected, archive.FittestItems[i].Fitness, "wrong item fitness at: %d", i)
	}

	// test update over allowed size
	fitness := 0.6
	org, err = genetics.NewOrganism(fitness, gen, 1)
	require.NoError(t, err, "failed to create new organism")
	err = archive.UpdateFittestWithOrganism(fillOrganismData(org, 0.0))
	require.NoError(t, err, "failed to update fittest")
	require.Len(t, archive.FittestItems, fittestAllowedSize)

	assert.Equal(t, fitness, archive.FittestItems[0].Fitness, "The item with maximal fitness at wrong position")
}

func TestNoveltyArchive_addNoveltyItem(t *testing.T) {
	archive := NewNoveltyArchive(1.0, nil)
	gen, err := genetics.ReadGenome(strings.NewReader(gnome_str), 1)
	require.NoError(t, err, "failed to read genome")
	org, err := genetics.NewOrganism(0.1, gen, 1)
	require.NoError(t, err, "failed to create new organism")
	org = fillOrganismData(org, 0.0)

	// test add novelty item
	item := org.Data.Value.(*NoveltyItem)
	archive.addNoveltyItem(item)

	require.Len(t, archive.NovelItems, 1, "wrong novelty items number")
	assert.True(t, archive.NovelItems[0].added, "items was added")
	assert.Equal(t, archive.Generation, archive.NovelItems[0].Generation, "wrong generation")
	assert.Equal(t, 1, archive.itemsAddedInGeneration, "wrong novelty items number for generation")
}

func TestNoveltyArchive_EvaluateIndividual(t *testing.T) {
	rand.Seed(42)
	in, out, nmax := 3, 2, 5
	linkProb := 0.5
	conf := neat.NeatContext{
		CompatThreshold: 0.5,
		PopSize:         10,
	}
	pop, err := genetics.NewPopulationRandom(in, out, nmax, false, linkProb, &conf)
	require.NoError(t, err, "failed to create population")
	require.NotNil(t, pop, "population expected")

	for i := 0; i < len(pop.Organisms); i++ {
		pop.Organisms[i].Fitness = 0.1 * (1.0 + float64(i))
		fillOrganismData(pop.Organisms[i], 0.1*(1.0+float64(i)))
	}

	metric := func(x, y *NoveltyItem) float64 {
		return (x.Fitness - y.Fitness) * (x.Fitness - y.Fitness)
	}
	archive := NewNoveltyArchive(1.0, metric)
	archive.Generation = 2

	// test evaluate only in archive
	//
	org := pop.Organisms[0]
	archive.EvaluateIndividualNovelty(org, pop, false)
	require.Len(t, archive.NovelItems, 1, "wrong novelty items number")
	assert.True(t, archive.NovelItems[0].added, "items was added")
	// check that data object properly filled
	item := org.Data.Value.(*NoveltyItem)
	assert.True(t, item.added)
	assert.Equal(t, archive.Generation, item.Generation)

	// test evaluate in population as well
	//
	archive.EvaluateIndividualNovelty(org, pop, true)
	require.Len(t, archive.NovelItems, 1, "wrong novelty items number")
	assert.NotEqual(t, 0.1, org.Fitness, "The organism fitness should be different from initial (0.1)")
}

func TestNoveltyArchive_EvaluatePopulation(t *testing.T) {
	rand.Seed(42)
	in, out, nmax := 3, 2, 5
	linkProb := 0.5
	conf := neat.NeatContext{
		CompatThreshold: 0.5,
		PopSize:         10,
	}
	pop, err := genetics.NewPopulationRandom(in, out, nmax, false, linkProb, &conf)
	require.NoError(t, err, "failed to create population")
	require.NotNil(t, pop, "population expected")

	for i := 0; i < len(pop.Organisms); i++ {
		pop.Organisms[i].Fitness = 0.1 * (1.0 + float64(i))
		fillOrganismData(pop.Organisms[i], 0.1*(1.0+float64(i)))
	}

	metric := func(x, y *NoveltyItem) float64 {
		return (x.Fitness - y.Fitness) * (x.Fitness - y.Fitness)
	}
	archive := NewNoveltyArchive(0.1, metric)
	archive.Generation = 2

	// test update fitness scores
	//
	archive.EvaluatePopulationNovelty(pop, true)
	for i := 0; i < len(pop.Organisms); i++ {
		notExpected := 0.1 * (1.0 + float64(i))
		assert.NotEqual(t, notExpected, pop.Organisms[i].Fitness, "Organism #%d fitness should be updated", i)
	}

	// test add to archive
	archive.EvaluatePopulationNovelty(pop, false)
	assert.Len(t, archive.NovelItems, 3, "wrong NovelItems count in the archive")
}

func fillOrganismData(org *genetics.Organism, novelty float64) *genetics.Organism {
	ni := NoveltyItem{
		Generation: org.Generation,
		Fitness:    org.Fitness,
		Novelty:    novelty,
	}
	org.Data = &genetics.OrganismData{Value: &ni}
	return org
}
