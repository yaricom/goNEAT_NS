package maze

import (
	"fmt"
	"github.com/yaricom/goNEAT/experiments"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT_NS/neatns"
	"os"
)

// The maze solving experiment with objective fitness-based optimization of NEAT algorithm
type MazeObjectiveEvaluator struct {
	// The output path to store execution results
	OutputPath string
	// The maze seed environment
	Environment *Environment

	// The target number of species to be maintained
	NumSpeciesTarget int
	// The species compatibility threshold adjustment frequency
	CompatAdjustFreq int
}

// Invoked before new trial run started
func (e MazeObjectiveEvaluator) TrialRunStarted(trial *experiments.Trial) {
	trialSim = mazeSimResults{
		trialID: trial.Id,
		records: new(RecordStore),
		archive: neatns.NewNoveltyArchive(archiveThresh, noveltyMetric),
	}
}

// This method evaluates one epoch for given population and prints results into output directory if any.
func (e MazeObjectiveEvaluator) GenerationEvaluate(pop *genetics.Population, epoch *experiments.Generation, context *neat.NeatContext) (err error) {
	// Evaluate each organism on a test
	for _, org := range pop.Organisms {
		res, err := e.orgEvaluate(org, pop, epoch)
		if err != nil {
			return err
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

	// Only print to file every print_every generations
	if epoch.Solved || epoch.Id%context.PrintEvery == 0 || epoch.Id == context.NumGenerations-1 {
		popPath := fmt.Sprintf("%s/gen_%d", experiments.OutDirForTrial(e.OutputPath, trialSim.trialID), epoch.Id)
		file, err := os.Create(popPath)
		if err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
		} else {
			pop.WriteBySpecies(file)
		}
	}

	if epoch.Solved {
		// print winner organism
		for _, org := range pop.Organisms {
			if org.IsWinner {
				// Prints the winner organism to file!
				orgPath := fmt.Sprintf("%s/%s_%d-%d", experiments.OutDirForTrial(e.OutputPath, trialSim.trialID),
					"maze_obj_winner", org.Phenotype.NodeCount(), org.Phenotype.LinkCount())
				file, err := os.Create(orgPath)
				if err != nil {
					neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism genome, reason: %s\n", err))
				} else if err = org.Genotype.Write(file); err != nil {
					neat.ErrorLog("Failed to save genotype")
				} else {
					neat.InfoLog(fmt.Sprintf("Generation #%d winner dumped to: %s\n", epoch.Id, orgPath))
				}
				break
			}
		}
		// store recorded data points and novelty archive
		e.storeRecorded()
	} else if epoch.Id == context.NumGenerations-1 {
		// the last epoch executed
		e.storeRecorded()
	} else {
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

	return err
}

func (e *MazeObjectiveEvaluator) storeRecorded() {
	// store recorded agents' performance
	recPath := fmt.Sprintf("%s/record.dat", experiments.OutDirForTrial(e.OutputPath, trialSim.trialID))
	recFile, err := os.Create(recPath)
	if err == nil {
		err = trialSim.records.Write(recFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to store agents' data records, reason: %s\n", err))
	}

	// print novelty points with maximal fitness
	npPath := fmt.Sprintf("%s/fittest_archive_points.txt", experiments.OutDirForTrial(e.OutputPath, trialSim.trialID))
	npFile, err := os.Create(npPath)
	if err == nil {
		err = trialSim.archive.PrintFittest(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print fittest  points from archive, reason: %s\n", err))
	}
}

// Evaluates individual organism against maze environment and returns true if organism was able to solve maze by navigating to exit
func (e *MazeObjectiveEvaluator) orgEvaluate(org *genetics.Organism, _ *genetics.Population, epoch *experiments.Generation) (bool, error) {
	// create record to store simulation results for organism
	record := AgentRecord{Generation: epoch.Id, AgentID: trialSim.individualsCounter}
	record.SpeciesID = org.Species.Id
	record.SpeciesAge = org.Species.Age

	// evaluate individual organism and get novelty point holding simulation results
	nItem, solved, err := mazeSimulationEvaluate(e.Environment, org, &record, nil)
	if err != nil {
		return false, err
	}
	nItem.IndividualID = org.Genotype.Id
	// assign organism fitness based on simulation results - the normalized distance between agent and maze exit
	org.Fitness = nItem.Fitness
	org.IsWinner = solved         // store if maze was solved
	org.Error = 1 - nItem.Fitness // error value consider how far  we are from exit normalized to (0;1] range

	if solved {
		// run simulation to store solver path
		pathPoints := make([]Point, e.Environment.TimeSteps)
		_, _, err := mazeSimulationEvaluate(e.Environment, org, &record, pathPoints)
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

	// update fittest organisms list - needed for debugging output
	org.Data = &genetics.OrganismData{Value: nItem} // store novelty item within organism data to avoid errors next
	if err = trialSim.archive.UpdateFittestWithOrganism(org); err != nil {
		return false, err
	}

	return solved, nil
}
