package maze

import (
	"fmt"
	"github.com/yaricom/goNEAT/v2/experiment"
	"github.com/yaricom/goNEAT/v2/experiment/utils"
	"github.com/yaricom/goNEAT/v2/neat"
	"github.com/yaricom/goNEAT/v2/neat/genetics"
	"github.com/yaricom/goNEAT_NS/v2/neatns"
	"math"
	"os"
)

// The initial novelty threshold for Novelty Archive
const archiveThresh = 6.0

// the novelty metric function for maze simulation
var noveltyMetric neatns.NoveltyMetric = func(x, y *neatns.NoveltyItem) float64 {
	diff := histDiff(x.Data, y.Data)
	return diff
}

// NewNoveltySearchEvaluator allows creating maze solving agent based on Novelty Search optimization.
// It will use provided MazeEnv to run simulation of the maze environment. The numSpeciesTarget specifies the
// target number of species to maintain in the population. If the number of species differ from the numSpeciesTarget it
// will be automatically adjusted with compatAdjustFreq frequency, i.e., at each epoch % compatAdjustFreq == 0
func NewNoveltySearchEvaluator(out string, mazeEnv *Environment, numSpeciesTarget, compatAdjustFreq int) (experiment.GenerationEvaluator, experiment.TrialRunObserver) {
	evaluator := &noveltySearchEvaluator{
		OutputPath:       out,
		MazeEnv:          mazeEnv,
		NumSpeciesTarget: numSpeciesTarget,
		CompatAdjustFreq: compatAdjustFreq,
	}
	return evaluator, evaluator
}

// noveltySearchEvaluator the maze solving experiment with Novelty Search optimization of NEAT algorithm
type noveltySearchEvaluator struct {
	// The output path to store execution results
	OutputPath string
	// The maze seed environment
	MazeEnv *Environment

	// The target number of species to be maintained
	NumSpeciesTarget int
	// The species compatibility threshold adjustment frequency
	CompatAdjustFreq int
}

func (e noveltySearchEvaluator) TrialRunStarted(trial *experiment.Trial) {
	opts := neatns.DefaultNoveltyArchiveOptions()
	opts.KNNNoveltyScore = 10
	trialSim = mazeSimResults{
		trialID: trial.Id,
		records: new(RecordStore),
		archive: neatns.NewNoveltyArchive(archiveThresh, noveltyMetric, opts),
	}
}

func (e noveltySearchEvaluator) TrialRunFinished(_ *experiment.Trial) {
	// the last epoch executed
	e.storeRecorded()
}

func (e noveltySearchEvaluator) EpochEvaluated(_ *experiment.Trial, _ *experiment.Generation) {
	// just stub
}

// GenerationEvaluate this method evaluates one epoch for given population and prints results into output directory if any.
func (e noveltySearchEvaluator) GenerationEvaluate(pop *genetics.Population, epoch *experiment.Generation, context *neat.Options) error {
	// Evaluate each organism on a test
	for i, org := range pop.Organisms {
		res, err := e.orgEvaluate(org, pop, epoch)
		if err != nil {
			return err
		}
		// store fitness based on objective proximity for statistical purposes
		if org.Data == nil {
			neat.ErrorLog(fmt.Sprintf("Novelty point not found at organism: %s", org))
			pop.Organisms[i].Fitness = 0.0
		} else {
			pop.Organisms[i].Fitness = org.Data.Value.(*neatns.NoveltyItem).Fitness
		}

		if res && (epoch.Best == nil || org.Fitness > epoch.Best.Fitness) {
			epoch.Solved = true
			epoch.WinnerNodes = len(org.Genotype.Nodes)
			epoch.WinnerGenes = org.Genotype.Extrons()
			epoch.WinnerEvals = trialSim.individualsCounter
			epoch.Best = org
		}
	}

	// Fill statistics about current epoch
	epoch.FillPopulationStatistics(pop)

	// Only print to file every print_every generation
	if epoch.Solved || epoch.Id%context.PrintEvery == 0 || epoch.Id == context.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.OutputPath, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	}

	if epoch.Solved {
		// print winner organism
		org := epoch.Best
		if depth, err := org.Phenotype.MaxActivationDepthFast(0); err == nil {
			neat.InfoLog(fmt.Sprintf("Activation depth of the winner: %d\n", depth))
		}

		genomeFile := "mazens_winner"
		// Prints the winner organism's Genome to the file!
		if orgPath, err := utils.WriteGenomePlain(genomeFile, e.OutputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's genome, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's genome dumped to: %s\n", epoch.Id, orgPath))
		}

		// Prints the winner organism's Phenotype to the Cytoscape JSON file!
		if orgPath, err := utils.WriteGenomeCytoscapeJSON(genomeFile, e.OutputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's phenome Cytoscape JSON graph, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's phenome Cytoscape JSON graph dumped to: %s\n",
				epoch.Id, orgPath))
		}
	} else if epoch.Id < context.NumGenerations-1 {
		// adjust archive settings
		trialSim.archive.EndOfGeneration()
		// refresh generation's novelty scores
		trialSim.archive.EvaluatePopulationNovelty(pop, true)

		speciesCount := len(pop.Species)

		// adjust species count by keeping it constant
		if epoch.Id%e.CompatAdjustFreq == 0 {
			if speciesCount < e.NumSpeciesTarget {
				context.CompatThreshold -= 0.1
			} else if speciesCount > e.NumSpeciesTarget {
				context.CompatThreshold += 0.1
			}

			// to avoid dropping too low
			if context.CompatThreshold < 0.3 {
				context.CompatThreshold = 0.3
			}
		}

		neat.InfoLog(fmt.Sprintf("%d species -> %d organisms [compatibility threshold: %.1f, target: %d]\n",
			speciesCount, len(pop.Organisms), context.CompatThreshold, e.NumSpeciesTarget))
	}

	return nil
}

func (e *noveltySearchEvaluator) storeRecorded() {
	// store recorded agents' performance
	recPath := fmt.Sprintf("%s/record.dat", utils.CreateOutDirForTrial(e.OutputPath, trialSim.trialID))
	recFile, err := os.Create(recPath)
	if err == nil {
		err = trialSim.records.Write(recFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to store agents' data records, reason: %s\n", err))
	}

	// print collected novelty points from archive
	npPath := fmt.Sprintf("%s/novelty_archive_points.txt", utils.CreateOutDirForTrial(e.OutputPath, trialSim.trialID))
	npFile, err := os.Create(npPath)
	if err == nil {
		err = trialSim.archive.PrintNoveltyPoints(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print novelty points from archive, reason: %s\n", err))
	}

	// print novelty points with maximal fitness
	npPath = fmt.Sprintf("%s/fittest_novelty_archive_points.txt", utils.CreateOutDirForTrial(e.OutputPath, trialSim.trialID))
	npFile, err = os.Create(npPath)
	if err == nil {
		err = trialSim.archive.PrintFittest(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print fittest novelty points from archive, reason: %s\n", err))
	}
}

// Evaluates individual organism against maze environment and returns true if organism was able to solve maze by navigating to exit
func (e *noveltySearchEvaluator) orgEvaluate(org *genetics.Organism, pop *genetics.Population, epoch *experiment.Generation) (bool, error) {
	// create record to store simulation results for organism
	record := AgentRecord{Generation: epoch.Id, AgentID: trialSim.individualsCounter}
	record.SpeciesID = org.Species.Id
	record.SpeciesAge = org.Species.Age

	// evaluate individual organism and get novelty point
	nItem, solved, err := mazeSimulationEvaluate(e.MazeEnv, org, &record, nil)
	if err != nil {
		if err == ErrOutputIsNaN {
			// corrupted genome, but OK to continue evolutionary process
			return false, nil
		}
		return false, err
	}
	nItem.IndividualID = org.Genotype.Id
	org.Data = &genetics.OrganismData{Value: nItem} // store novelty item within organism data
	org.IsWinner = solved                           // store if maze was solved
	org.Error = 1 - nItem.Fitness                   // error value consider how far  we are from exit normalized to (0;1] range

	// calculate novelty of new individual within archive of known novel items
	if !solved {
		trialSim.archive.EvaluateIndividualNovelty(org, pop, false)
		record.Novelty = org.Data.Value.(*neatns.NoveltyItem).Novelty // put it to the record
	} else {
		// solution found - set to maximal possible value
		record.Novelty = math.MaxFloat64

		// run simulation to store solver path
		pathPoints := make([]Point, e.MazeEnv.TimeSteps)
		if _, _, err = mazeSimulationEvaluate(e.MazeEnv, org, nil, pathPoints); err != nil {
			neat.ErrorLog("Solver's path simulation failed\n")
			return false, err
		}
		trialSim.records.SolverPathPoints = pathPoints
	}

	// add record
	trialSim.records.Records = append(trialSim.records.Records, record)

	// increment tested unique individuals counter
	trialSim.individualsCounter++

	// update fittest organisms list
	if err = trialSim.archive.UpdateFittestWithOrganism(org); err != nil {
		return false, err
	}
	return solved, nil
}
