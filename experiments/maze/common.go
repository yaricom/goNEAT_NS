// The maze solving experiments based on NEAT methodology with Novelty Search and Fitness based optimization
package maze

import (
	"fmt"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT_NS/neatns"
	"math"
	"github.com/yaricom/goNEAT/neat/network"
)

// The simulation results for one trial
var trialSim mazeSimResults

// The structure to hold maze simulator evaluation results
type mazeSimResults struct {
	// The record store for evaluated agents
	records        *RecordStore
	// The novelty archive
	archive        *neatns.NoveltyArchive

	// The current trial
	trialID        int
	// The evaluated individuals counter within current trial
	individCounter int
}

// calculates item-wise difference between two vectors
func histDiff(in1, in2 []float64) float64 {
	size := len(in1)
	diff_accum := 0.0
	for i := 0; i < size; i++ {
		diff := in1[i] - in2[i]
		diff_accum += math.Abs(diff)
	}
	return diff_accum / float64(size)
}


// To evaluate an individual organism within provided maze environment and to create corresponding novelty point.
// If maze was solved during simulation the second returned parameter will be true.
func mazeSimulationEvaluate(env *Environment, org *genetics.Organism, record *AgentRecord, pathPoints []Point) (*neatns.NoveltyItem, bool, error) {
	n_item := neatns.NewNoveltyItem()

	// initialize maze simulation's environment specific to the provided organism - this will be a copy
	// of primordial environment provided
	org_env, err := mazeSimulationInit(*env, org)
	if err != nil {
		return nil, false, err
	}

	// do specified amount of time steps emulations or while exit not found
	steps := 0
	for i := 0; i < org_env.TimeSteps && !org_env.ExitFound; i++ {
		err := mazeSimulationStep(org_env, org)
		if err != nil {
			return nil, false, err
		}
		// store agent path points at given sample size
		if (org_env.TimeSteps - i) % org_env.SampleSize == 0 {
			n_item.Data = append(n_item.Data, org_env.Hero.Location.X)
			n_item.Data = append(n_item.Data, org_env.Hero.Location.Y)
		}

		// store all path points if requested
		if pathPoints != nil {
			pathPoints[i] = org_env.Hero.Location
		}
		steps++
	}

	if org_env.ExitFound {
		neat.InfoLog(fmt.Sprintf("Maze solved in: %d steps\n", steps))
	}

	// calculate fitness of an organism as closeness to target
	fitness := org_env.AgentDistanceToExit()

	// normalize fitness value in range (0;1] and store it
	fitness = (env.initialDistance - fitness) / env.initialDistance
	if fitness <= 0 {
		fitness = 0.01
	}

	n_item.Fitness = fitness

	// store final agent coordinates as organism's novelty characteristics
	n_item.Data = append(n_item.Data, org_env.Hero.Location.X)
	n_item.Data = append(n_item.Data, org_env.Hero.Location.Y)

	if record != nil {
		record.Fitness = fitness
		record.X = org_env.Hero.Location.X
		record.Y = org_env.Hero.Location.Y
		record.GotExit = org_env.ExitFound
	}

	return n_item, org_env.ExitFound, nil
}


// To initialize the maze simulation within provided environment copy and for given organism.
// Returns new environment for simulation against given organism
func mazeSimulationInit(env Environment, org *genetics.Organism) (*Environment, error) {

	// get Organism phenotype's network depth
	net_depth, err := org.Phenotype.MaxDepth() // The max depth of the network to be activated
	if err != nil {
		if err != network.NetErrDepthCalculationFailedLoopDetected {
			return nil, err
		} else {
			neat.InfoLog(fmt.Sprintf("Network loop detected, using default depth: %d for organism: %d\n",
				net_depth, org.Genotype.Id))
		}
	}
	neat.DebugLog(fmt.Sprintf("Network depth: %d for organism: %d\n", net_depth, org.Genotype.Id))
	if net_depth == 0 {
		neat.DebugLog(fmt.Sprintf("ALERT: Network depth is ZERO for Genome: %s", org.Genotype))
	}

	// flush the neural net
	org.Phenotype.Flush()
	// update the maze
	err = env.Update()
	if err != nil {
		neat.ErrorLog("Failed to update environment")
		return nil, err
	}

	// create neural net inputs from environment
	inputs, err := env.GetInputs()
	if err != nil {
		return nil, err
	}

	// load into neural net
	if err = org.Phenotype.LoadSensors(inputs); err != nil {
		return nil, err
	}

	// propagate input through the phenotype net

	// Relax phenotype net and get output
	if _, err = org.Phenotype.Activate();err != nil && err != network.NetErrExceededMaxActivationAttempts {
		neat.ErrorLog("Failed to activate network init 1")
		return nil, err
	}

	// use depth to ensure relaxation at each layer
	for relax := 0; relax <= net_depth; relax++ {
		if _, err = org.Phenotype.Activate();err != nil && err != network.NetErrExceededMaxActivationAttempts {
			neat.ErrorLog("Failed to activate network init 2")
			return nil, err
		}
	}

	return &env, nil
}

// To execute a time step of the maze simulation evaluation within given Environment for provided Organism
func mazeSimulationStep(env *Environment, org *genetics.Organism) error {
	// get simulation parameters as inputs to organism's network
	inputs, err := env.GetInputs()
	if err != nil {
		return err
	}
	org.Phenotype.LoadSensors(inputs)
	_, err = org.Phenotype.Activate()
	if err != nil && err != network.NetErrExceededMaxActivationAttempts {
		neat.ErrorLog("Failed to activate network simulation")
		return err
	}

	// use the net's outputs to change heading and velocity of maze agent
	err = env.ApplyOutputs(org.Phenotype.Outputs[0].Activation, org.Phenotype.Outputs[1].Activation)
	if err != nil {
		neat.ErrorLog("Failed to apply outputs")
		return err
	}

	// update the environment
	err = env.Update()
	if err != nil {
		neat.ErrorLog("Failed to update environment")
		return err
	}

	return nil
}