package maze

import (
	"github.com/stretchr/testify/assert"
	"github.com/yaricom/goNEAT/v3/neat"
	"testing"
)

func TestCommon_adjustSpeciesNumber(t *testing.T) {
	initialThreshold := 0.5
	options := &neat.Options{
		CompatThreshold: initialThreshold,
	}

	// check no changes
	epochID := 1
	adjustFrequency := 5
	speciesCount := 10
	numberSpeciesTarget := 20
	adjustSpeciesNumber(speciesCount, epochID, adjustFrequency, numberSpeciesTarget, options)
	assert.Equal(t, initialThreshold, options.CompatThreshold, "no change expected")

	// check speciesCount < numberSpeciesTarget
	epochID = adjustFrequency
	adjustSpeciesNumber(speciesCount, epochID, adjustFrequency, numberSpeciesTarget, options)
	assert.Equal(t, initialThreshold-compatibilityThresholdStep, options.CompatThreshold)

	// check speciesCount > numberSpeciesTarget
	options.CompatThreshold = initialThreshold
	speciesCount = numberSpeciesTarget + 1
	adjustSpeciesNumber(speciesCount, epochID, adjustFrequency, numberSpeciesTarget, options)
	assert.Equal(t, initialThreshold+compatibilityThresholdStep, options.CompatThreshold)

	// check speciesCount == numberSpeciesTarget
	options.CompatThreshold = initialThreshold
	speciesCount = numberSpeciesTarget
	adjustSpeciesNumber(speciesCount, epochID, adjustFrequency, numberSpeciesTarget, options)
	assert.Equal(t, initialThreshold, options.CompatThreshold)

	// check avoiding of dropping too low
	options.CompatThreshold = compatibilityThresholdMinValue
	speciesCount = numberSpeciesTarget - 1
	adjustSpeciesNumber(speciesCount, epochID, adjustFrequency, numberSpeciesTarget, options)
	assert.Equal(t, compatibilityThresholdMinValue, options.CompatThreshold)
}
