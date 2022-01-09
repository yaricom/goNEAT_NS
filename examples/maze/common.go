// Package maze provides the maze solving experiments based on NEAT methodology with Novelty Search and Fitness
// based optimization.
package maze

import (
	"fmt"
	"github.com/yaricom/goNEAT/v2/neat"
	"github.com/yaricom/goNEAT/v2/neat/genetics"
	"github.com/yaricom/goNEAT/v2/neat/network"
	"github.com/yaricom/goNEAT_NS/v2/neatns"
	"math"
)

// The simulation results for one trial
var trialSim mazeSimResults

// The structure to hold maze simulator evaluation results
type mazeSimResults struct {
	// The record store for evaluated agents
	records *RecordStore
	// The novelty archive
	archive *neatns.NoveltyArchive

	// The current trial
	trialID int
	// The evaluated individuals counter within current trial
	individualsCounter int
}

// calculates item-wise difference between two vectors
func histDiff(left, right []float64) float64 {
	size := len(left)
	diffAccum := 0.0
	for i := 0; i < size; i++ {
		diff := left[i] - right[i]
		diffAccum += math.Abs(diff)
	}
	return diffAccum / float64(size)
}

// To evaluate an individual organism within provided maze environment and to create corresponding novelty point.
// If maze was solved during simulation the second returned parameter will be true.
func mazeSimulationEvaluate(env *Environment, org *genetics.Organism, record *AgentRecord, pathPoints []Point) (*neatns.NoveltyItem, bool, error) {
	nItem := neatns.NewNoveltyItem()

	// get Organism phenotype's network depth
	netDepth, err := org.Phenotype.MaxActivationDepthFast(1) // The max depth of the network to be activated
	if err != nil {
		neat.WarnLog(fmt.Sprintf(
			"Failed to estimate maximal depth of the network:\n%s\nUsing default depth: %d",
			org.Genotype, netDepth))
	}
	neat.DebugLog(fmt.Sprintf("Network depth: %d for organism: %d\n", netDepth, org.Genotype.Id))
	if netDepth == 0 {
		neat.DebugLog(fmt.Sprintf("ALERT: Network depth is ZERO for Genome: %s", org.Genotype))
	}

	// initialize maze simulation's environment specific to the provided organism - this will be a copy
	// of primordial environment provided
	orgEnv, err := mazeSimulationInit(*env, org, netDepth)
	if err != nil {
		return nil, false, err
	}

	// do a specified amount of time steps emulations or while exit not found
	steps := 0
	for i := 0; i < orgEnv.TimeSteps && !orgEnv.ExitFound; i++ {
		if err = mazeSimulationStep(orgEnv, org, netDepth); err != nil {
			return nil, false, err
		}
		// store agent path points at given sample size
		if (orgEnv.TimeSteps-i)%orgEnv.SampleSize == 0 {
			nItem.Data = append(nItem.Data, orgEnv.Hero.Location.X)
			nItem.Data = append(nItem.Data, orgEnv.Hero.Location.Y)
		}

		// store all path points if requested
		if pathPoints != nil {
			pathPoints[i] = orgEnv.Hero.Location
		}
		steps++
	}

	if orgEnv.ExitFound {
		neat.InfoLog(fmt.Sprintf("Maze solved in: %d steps\n", steps))
	}

	// calculate fitness of an organism as closeness to target
	fitness := orgEnv.AgentDistanceToExit()

	// normalize fitness value in range (0;1] and store it
	fitness = (env.initialDistance - fitness) / env.initialDistance
	if fitness <= 0 {
		fitness = 0.01
	}

	nItem.Fitness = fitness

	// store final agent coordinates as organism's novelty characteristics
	nItem.Data = append(nItem.Data, orgEnv.Hero.Location.X)
	nItem.Data = append(nItem.Data, orgEnv.Hero.Location.Y)

	if record != nil {
		record.Fitness = fitness
		record.X = orgEnv.Hero.Location.X
		record.Y = orgEnv.Hero.Location.Y
		record.GotExit = orgEnv.ExitFound
	}

	return nItem, orgEnv.ExitFound, nil
}

// To initialize the maze simulation within provided environment copy and for given organism.
// Returns new environment for simulation against given organism
func mazeSimulationInit(env Environment, org *genetics.Organism, netDepth int) (*Environment, error) {
	// flush the neural net
	if _, err := org.Phenotype.Flush(); err != nil {
		neat.ErrorLog("Failed to flush phenotype")
		return nil, err
	}
	// update the maze
	if err := env.Update(); err != nil {
		neat.ErrorLog("Failed to update environment")
		return nil, err
	}

	// create neural net inputs from environment
	if inputs, err := env.GetInputs(); err != nil {
		return nil, err
	} else if err = org.Phenotype.LoadSensors(inputs); err != nil { // load into neural net
		return nil, err
	}

	// propagate input through the phenotype net

	// Use depth to ensure full relaxation
	if _, err := org.Phenotype.ForwardSteps(netDepth); err != nil && err != network.ErrNetExceededMaxActivationAttempts {
		neat.ErrorLog(fmt.Sprintf("Failed to activate network at simulation init: %s", err))
		return nil, err
	}

	return &env, nil
}

// To execute a time step of the maze simulation evaluation within given Environment for provided Organism
func mazeSimulationStep(env *Environment, org *genetics.Organism, netDepth int) error {
	// get simulation parameters as inputs to organism's network
	if inputs, err := env.GetInputs(); err != nil {
		return err
	} else if err = org.Phenotype.LoadSensors(inputs); err != nil {
		neat.ErrorLog("Failed to load sensors")
		return err
	}
	if _, err := org.Phenotype.ForwardSteps(netDepth); err != nil && err != network.ErrNetExceededMaxActivationAttempts {
		neat.ErrorLog(fmt.Sprintf("Failed to activate network at simulation init: %s", err))
		return err
	}

	// use the net's outputs to change heading and velocity of maze agent
	if err := env.ApplyOutputs(org.Phenotype.Outputs[0].Activation, org.Phenotype.Outputs[1].Activation); err != nil {
		neat.ErrorLog("Failed to apply outputs")
		return err
	}

	// update the environment
	if err := env.Update(); err != nil {
		neat.ErrorLog("Failed to update environment")
		return err
	}

	return nil
}
