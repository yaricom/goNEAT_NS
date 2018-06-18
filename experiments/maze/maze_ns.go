package maze

import (
	"github.com/yaricom/goNEAT_NS/neatns"
	"github.com/yaricom/goNEAT_NS/experiments"
)


// the novelty metric function for maze simulation
var noveltyMetric neatns.NoveltyMetric = func(x, y *neatns.NoveltyItem) float64 {
	diff := 0.0
	for i := 0; i < len(x.Data); i++ {
		diff += experiments.HistDiff(x.Data[i], y.Data[i])
	}
	return diff
}

