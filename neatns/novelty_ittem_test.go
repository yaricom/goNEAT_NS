package neatns

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoveltyItem_String(t *testing.T) {
	item := NoveltyItem{
		Generation:   1,
		IndividualID: 10,
		Fitness:      0.5,
		Novelty:      25.35,
		Age:          2,
		Data:         []float64{100.1, 123.9},
	}

	str := item.String()
	expected := "Novelty: 25.35 Fitness: 0.500000 Generation: 1 Individual: 10\n\tPoint:  100.100 123.900"
	assert.Equal(t, expected, str)
}
