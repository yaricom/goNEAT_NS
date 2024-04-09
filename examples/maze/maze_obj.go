package maze

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/experiment/utils"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT_NS/v4/neatns"
	"os"
)

// NewMazeObjectiveEvaluator allows creating maze solving agent based on Novelty Search optimization.
// It will use provided MazeEnv to run simulation of the maze environment. The numSpeciesTarget specifies the
// target number of species to maintain in the population. If the number of species differ from the numSpeciesTarget it
// will be automatically adjusted with compatAdjustFreq frequency, i.e., at each epoch % compatAdjustFreq == 0
func NewMazeObjectiveEvaluator(out string, mazeEnv *Environment, numSpeciesTarget, compatAdjustFreq int) (experiment.GenerationEvaluator, experiment.TrialRunObserver) {
	evaluator := &objectiveEvaluator{
		outputPath:       out,
		mazeEnv:          mazeEnv,
		numSpeciesTarget: numSpeciesTarget,
		compatAdjustFreq: compatAdjustFreq,
	}
	return evaluator, evaluator
}

// objectiveEvaluator The maze solving experiment evaluator with objective fitness-based optimization of NEAT algorithm
type objectiveEvaluator struct {
	// The output path to store execution results
	outputPath string
	// The maze seed environment
	mazeEnv *Environment

	// The target number of species to be maintained
	numSpeciesTarget int
	// The species compatibility threshold adjustment frequency
	compatAdjustFreq int
}

func (e *objectiveEvaluator) TrialRunStarted(trial *experiment.Trial) {
	trialSim = mazeSimResults{
		trialID: trial.Id,
		records: new(RecordStore),
		archive: neatns.NewNoveltyArchive(archiveThresh, NoveltyMetric, neatns.DefaultNoveltyArchiveOptions()),
	}
}

func (e *objectiveEvaluator) TrialRunFinished(_ *experiment.Trial) {
	// the last epoch executed
	e.storeRecorded()
}

func (e *objectiveEvaluator) EpochEvaluated(_ *experiment.Trial, _ *experiment.Generation) {
	// just stub
}

// GenerationEvaluate evaluates one epoch for given population and prints results into output directory if any.
func (e *objectiveEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	options, ok := neat.FromContext(ctx)
	if !ok {
		return neat.ErrNEATOptionsNotFound
	}
	// Evaluate each organism on a test
	for _, org := range pop.Organisms {
		res, err := e.orgEvaluate(org, pop, epoch)
		if err != nil {
			return err
		}
		if res && (epoch.Champion == nil || org.Fitness > epoch.Champion.Fitness) {
			epoch.Solved = true
			epoch.WinnerNodes = len(org.Genotype.Nodes)
			epoch.WinnerGenes = org.Genotype.Extrons()
			epoch.WinnerEvals = trialSim.individualsCounter
			epoch.Champion = org
		}
	}

	// Fill statistics about current epoch
	epoch.FillPopulationStatistics(pop)

	// Only print to file every print_every generation
	if epoch.Solved || epoch.Id%options.PrintEvery == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputPath, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	}

	if epoch.Solved {
		// print winner organism
		org := epoch.Champion
		utils.PrintActivationDepth(org, true)

		genomeFile := "maze_obj_winner"
		// Prints the winner organism to file!
		if orgPath, err := utils.WriteGenomePlain(genomeFile, e.outputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's genome, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's genome dumped to: %s\n", epoch.Id, orgPath))
		}

		// Prints the winner organism's Phenotype to the Cytoscape JSON file!
		if orgPath, err := utils.WriteGenomeCytoscapeJSON(genomeFile, e.outputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's phenome Cytoscape JSON graph, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's phenome Cytoscape JSON graph dumped to: %s\n",
				epoch.Id, orgPath))
		}
	} else if epoch.Id < options.NumGenerations-1 {
		speciesCount := len(pop.Species)

		// adjust species count by keeping it constant
		adjustSpeciesNumber(speciesCount, epoch.Id, e.compatAdjustFreq, e.numSpeciesTarget, options)

		neat.InfoLog(fmt.Sprintf("%d species -> %d organisms [compatibility threshold: %.1f, target: %d]\n",
			speciesCount, len(pop.Organisms), options.CompatThreshold, e.numSpeciesTarget))
	}

	return nil
}

func (e *objectiveEvaluator) storeRecorded() {
	// store recorded agents' performance
	recPath := fmt.Sprintf("%s/record.dat", utils.CreateOutDirForTrial(e.outputPath, trialSim.trialID))
	recFile, err := os.Create(recPath)
	if err == nil {
		err = trialSim.records.Write(recFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to store agents' data records, reason: %s\n", err))
	}

	// print novelty points with maximal fitness
	npPath := fmt.Sprintf("%s/fittest_archive_points.json", utils.CreateOutDirForTrial(e.outputPath, trialSim.trialID))
	npFile, err := os.Create(npPath)
	if err == nil {
		err = trialSim.archive.DumpFittest(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print fittest  points from archive, reason: %s\n", err))
	}
}

// Evaluates individual organism against maze environment and returns true if organism was able to solve maze by navigating to exit
func (e *objectiveEvaluator) orgEvaluate(org *genetics.Organism, _ *genetics.Population, epoch *experiment.Generation) (bool, error) {
	// create record to store simulation results for organism
	record := AgentRecord{Generation: epoch.Id, AgentID: trialSim.individualsCounter}
	record.SpeciesID = org.Species.Id
	record.SpeciesAge = org.Species.Age

	// evaluate individual organism and get novelty point holding simulation results
	nItem, solved, err := mazeSimulationEvaluate(e.mazeEnv, org, &record, nil)
	if err != nil {
		if errors.Is(err, ErrOutputIsNaN) {
			// corrupted genome, but OK to continue evolutionary process
			return false, nil
		}
		return false, err
	}
	nItem.IndividualID = org.Genotype.Id
	// assign organism fitness based on simulation results - the normalized distance between agent and maze exit
	org.Fitness = nItem.Fitness
	org.IsWinner = solved         // store if maze was solved
	org.Error = 1 - nItem.Fitness // error value consider how far  we are from exit normalized to (0;1] range

	if solved {
		// run simulation to store solver path
		pathPoints := make([]Point, e.mazeEnv.TimeSteps)
		_, _, err := mazeSimulationEvaluate(e.mazeEnv, org, &record, pathPoints)
		if err != nil {
			neat.ErrorLog("Solver's path simulation failed\n")
			return false, err
		}
		trialSim.records.SolverPathPoints = pathPoints
	}

	// add record
	trialSim.records.Records = append(trialSim.records.Records, record)

	// increment tested unique individuals counter
	trialSim.individualsCounter++

	// update the fittest organisms list - needed for debugging output
	org.Data = &genetics.OrganismData{Value: nItem} // store novelty item within organism data to avoid errors next
	if err = trialSim.archive.UpdateFittestWithOrganism(org); err != nil {
		return false, err
	}

	return solved, nil
}
