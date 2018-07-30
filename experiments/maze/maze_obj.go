package maze

import (
	"github.com/yaricom/goNEAT/experiments"
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT/neat"
	"fmt"
	"github.com/yaricom/goNEAT_NS/neatns"
	"os"
)


// The maze solving experiment with objective fitness-based optimization of NEAT algorithm
type MazeObjectiveEvaluator struct {
	// The output path to store execution results
	OutputPath       string
	// The maze seed environment
	Environment      *Environment

	// The target number of species to be maintained
	NumSpeciesTarget int
	// The species compatibility threshold adjustment frequency
	CompatAdjustFreq int
}

// Invoked before new trial run started
func (ev MazeObjectiveEvaluator) TrialRunStarted(trial *experiments.Trial) {
	trialSim = mazeSimResults{
		trialID : trial.Id,
		records : new(RecordStore),
		archive : neatns.NewNoveltyArchive(archive_thresh, noveltyMetric),
	}
}

// This method evaluates one epoch for given population and prints results into output directory if any.
func (ev MazeObjectiveEvaluator) GenerationEvaluate(pop *genetics.Population, epoch *experiments.Generation, context *neat.NeatContext) (err error) {
	// Evaluate each organism on a test
	for _, org := range pop.Organisms {
		res, err := ev.orgEvaluate(org, pop, epoch)
		if err != nil {
			return err
		}
		if res {
			epoch.Solved = true
			epoch.WinnerNodes = len(org.Genotype.Nodes)
			epoch.WinnerGenes = org.Genotype.Extrons()
			epoch.WinnerEvals = trialSim.individCounter
			epoch.Best = org

			break // we have a winner
		}
	}

	// Fill statistics about current epoch
	epoch.FillPopulationStatistics(pop)

	// Only print to file every print_every generations
	if epoch.Solved || epoch.Id % context.PrintEvery == 0 || epoch.Id == context.NumGenerations - 1 {
		pop_path := fmt.Sprintf("%s/gen_%d", outDirForTrial(ev.OutputPath, trialSim.trialID), epoch.Id)
		file, err := os.Create(pop_path)
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
				org_path := fmt.Sprintf("%s/%s_%d-%d", outDirForTrial(ev.OutputPath, trialSim.trialID),
					"maze_obj_winner", org.Phenotype.NodeCount(), org.Phenotype.LinkCount())
				file, err := os.Create(org_path)
				if err != nil {
					neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism genome, reason: %s\n", err))
				} else {
					org.Genotype.Write(file)
					neat.InfoLog(fmt.Sprintf("Generation #%d winner dumped to: %s\n", epoch.Id, org_path))
				}
				break
			}
		}
		// store recorded data points and novelty archive
		ev.storeRecorded()
	} else if epoch.Id == context.NumGenerations - 1 {
		// the last epoch executed
		ev.storeRecorded()
	} else {
		speciesCount := len(pop.Species)

		// adjust species count by keeping it constant
		if epoch.Id % ev.CompatAdjustFreq == 0 {
			if speciesCount < ev.NumSpeciesTarget {
				context.CompatThreshold -= 0.1
			} else if speciesCount > ev.NumSpeciesTarget {
				context.CompatThreshold += 0.1
			}

			// to avoid dropping too low
			if context.CompatThreshold < 0.3 {
				context.CompatThreshold = 0.3
			}
		}

		neat.InfoLog(fmt.Sprintf("%d species -> %d organisms [compatibility threshold: %.1f, target: %d]\n",
			speciesCount, len(pop.Organisms), context.CompatThreshold, ev.NumSpeciesTarget))

		// Move to the next epoch if failed to find winner
		neat.DebugLog(">>>>> start next generation")
		_, err = pop.Epoch(epoch.Id + 1, context)
	}

	return err
}

func (ev *MazeObjectiveEvaluator) storeRecorded() {
	// store recorded agents' performance
	rec_path := fmt.Sprintf("%s/record.dat", outDirForTrial(ev.OutputPath, trialSim.trialID))
	rec_file, err := os.Create(rec_path)
	if err == nil {
		err = trialSim.records.Write(rec_file)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to store agents' data records, reason: %s\n", err))
	}

	// print novelty points with maximal fitness
	np_path := fmt.Sprintf("%s/fittest_archive_points.txt", outDirForTrial(ev.OutputPath, trialSim.trialID))
	np_file, err := os.Create(np_path)
	if err == nil {
		trialSim.archive.PrintFittest(np_file)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print fittest  points from archive, reason: %s\n", err))
	}
}

// Evaluates individual organism against maze environment and returns true if organism was able to solve maze by navigating to exit
func (ev *MazeObjectiveEvaluator) orgEvaluate(org *genetics.Organism, pop *genetics.Population, epoch *experiments.Generation) (bool, error) {
	// create record to store simulation results for organism
	record := AgentRecord{Generation:epoch.Id, AgentID:trialSim.individCounter}
	record.SpeciesID = org.Species.Id
	record.SpeciesAge = org.Species.Age

	// evaluate individual organism and get novelty point holding simulation results
	n_item, solved, err := mazeSimulationEvaluate(ev.Environment, org, &record)
	if err != nil {
		return false, err
	}
	n_item.IndividualID = org.Genotype.Id
	// assign organism fitness based on simulation results - the normalized distance between agent and maze exit
	org.Fitness = n_item.Fitness
	org.IsWinner = solved // store if maze was solved
	org.Error = 1 - n_item.Fitness // error value consider how far  we are from exit normalized to (0;1] range

	// add record
	trialSim.records.Records = append(trialSim.records.Records, record)

	// increment tested unique individuals counter
	trialSim.individCounter++

	// update fittest organisms list - needed for debugging output
	org.Data = &genetics.OrganismData{Value:n_item}  // store novelty item within organism data to avoid errors next
	trialSim.archive.UpdateFittestWithOrganism(org)

	return solved, nil
}
