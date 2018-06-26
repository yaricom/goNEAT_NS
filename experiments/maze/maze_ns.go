package maze

import (
	"github.com/yaricom/goNEAT_NS/neatns"
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/experiments"
)

// The initial novelty threshold for Novelty Archive
const archive_thresh = 6.0


// the novelty metric function for maze simulation
var noveltyMetric neatns.NoveltyMetric = func(x, y *neatns.NoveltyItem) float64 {
	diff := histDiff(x.Data, y.Data)
	return diff
}

// The maze solving experiment with Novelty Search optimization of NEAT algorithm
type MazeNoveltySearchEvaluator struct {
	// The output path to store execution results
	OutputPath     string
	// The maze seed environment
	environment    *Environment

	// The record store for evaluated agents
	store          *RecordStore
	// The novelty archive
	archive        *neatns.NoveltyArchive

	// The current trial
	trialID        int
	// The evaluated individuals counter within current trial
	individCounter int
}

// Invoked before new trial run started
func (ev *MazeNoveltySearchEvaluator) TrialRunStarted(trial *experiments.Trial) {
	ev.trialID = trial.Id
	ev.individCounter = 0

	// create new record store and novelty archive
	ev.store = new(RecordStore)
	ev.archive = neatns.NewNoveltyArchive(archive_thresh, noveltyMetric)
}

// This method evaluates one epoch for given population and prints results into output directory if any.
func (ev *MazeNoveltySearchEvaluator) GenerationEvaluate(pop *genetics.Population, epoch *experiments.Generation, context *neat.NeatContext) (err error) {

}