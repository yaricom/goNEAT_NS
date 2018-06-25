package maze

import (
	"github.com/yaricom/goNEAT_NS/neatns"
	"github.com/yaricom/goNEAT_NS/experiments"
)


// the novelty metric function for maze simulation
var noveltyMetric neatns.NoveltyMetric = func(x, y *neatns.NoveltyItem) float64 {
	diff := experiments.HistDiff(x.Data, y.Data)
	return diff
}

