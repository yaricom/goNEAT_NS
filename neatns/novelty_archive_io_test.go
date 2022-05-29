package neatns

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoveltyArchive_PrintFittest(t *testing.T) {
	pop, err := createRandomPopulation(3, 2, 5, 0.5)
	require.NoError(t, err, "failed to create population")
	require.NotNil(t, pop, "population expected")

	archive := NewNoveltyArchive(0.1, squareMetric, DefaultNoveltyArchiveOptions())
	archive.Generation = 2

	archive.EvaluatePopulationNovelty(pop, true)
	err = archive.UpdateFittestWithOrganism(pop.Organisms[0])
	require.NoError(t, err)

	var buf bytes.Buffer
	err = archive.DumpFittest(&buf)
	require.NoError(t, err)

	// decode and check
	var novelItems []*NoveltyItem
	err = json.Unmarshal(buf.Bytes(), &novelItems)

	assertItemsEqual(archive.FittestItems, novelItems, t)
}

func TestNoveltyArchive_PrintFittest_no_points(t *testing.T) {
	archive := NewNoveltyArchive(0.1, squareMetric, DefaultNoveltyArchiveOptions())

	var buf bytes.Buffer
	err := archive.DumpFittest(&buf)
	assert.Error(t, err, ErrNoFittestItems.Error())
}

func TestNoveltyArchive_PrintNoveltyPoints(t *testing.T) {
	pop, err := createRandomPopulation(3, 2, 5, 0.5)
	require.NoError(t, err, "failed to create population")
	require.NotNil(t, pop, "population expected")

	archive := NewNoveltyArchive(0.1, squareMetric, DefaultNoveltyArchiveOptions())
	archive.Generation = 2

	archive.EvaluatePopulationNovelty(pop, false)
	var buf bytes.Buffer
	err = archive.DumpNoveltyPoints(&buf)
	require.NoError(t, err)

	// decode and check
	var novelItems []*NoveltyItem
	err = json.Unmarshal(buf.Bytes(), &novelItems)

	assertItemsEqual(archive.NovelItems, novelItems, t)
}

func TestNoveltyArchive_PrintNoveltyPoints_no_points(t *testing.T) {
	archive := NewNoveltyArchive(0.1, squareMetric, DefaultNoveltyArchiveOptions())

	var buf bytes.Buffer
	err := archive.DumpNoveltyPoints(&buf)
	assert.Error(t, err, ErrNoNovelItems.Error())
}

func assertItemsEqual(expected, actual []*NoveltyItem, t *testing.T) {
	require.Equal(t, len(expected), len(actual))
	for i, ni := range expected {
		assert.Equal(t, ni.Age, actual[i].Age)
		assert.Equal(t, ni.Novelty, actual[i].Novelty)
		assert.Equal(t, ni.Fitness, actual[i].Fitness)
		assert.Equal(t, ni.Generation, actual[i].Generation)
		assert.Equal(t, ni.IndividualID, actual[i].IndividualID)
		assert.EqualValues(t, ni.Data, actual[i].Data)
	}
}
