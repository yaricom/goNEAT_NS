package neatns

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const fittestStr = "/* Novelty: 0.29 Fitness: 0.100000 Generation: 2 Individual: 0\n\tPoint:  0.100 */\n"
const novPointsStr = "/* Novelty: 0.00 Fitness: 0.100000 Generation: 2 Individual: 0\n\tPoint:  0.100 */\n/* Novelty: 0.16 Fitness: 0.500000 Generation: 2 Individual: 0\n\tPoint:  0.100 */\n/* Novelty: 0.16 Fitness: 0.900000 Generation: 2 Individual: 0\n\tPoint:  0.100 */\n"

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
	err = archive.PrintFittest(&buf)
	require.NoError(t, err)
	assert.Equal(t, fittestStr, buf.String())
	t.Logf(buf.String())
}

func TestNoveltyArchive_PrintNoveltyPoints(t *testing.T) {
	pop, err := createRandomPopulation(3, 2, 5, 0.5)
	require.NoError(t, err, "failed to create population")
	require.NotNil(t, pop, "population expected")

	archive := NewNoveltyArchive(0.1, squareMetric, DefaultNoveltyArchiveOptions())
	archive.Generation = 2

	archive.EvaluatePopulationNovelty(pop, false)
	var buf bytes.Buffer
	err = archive.PrintNoveltyPoints(&buf)
	require.NoError(t, err)
	assert.Equal(t, novPointsStr, buf.String())
	t.Logf(buf.String())
}
