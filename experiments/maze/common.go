// The maze solving experiments based on NEAT methodology with Novelty Search and Fitness based optimization
package maze

import (
	"fmt"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/neat/genetics"
)


// To initialize the maze simulation within provided environment copy and for given organism.
// Returns new environment for simulation against given organism
func mazeSimulationInit(env Environment, org *genetics.Organism) (*Environment, error) {

	// get Organism phenotype's network depth
	net_depth, err := org.Phenotype.MaxDepth() // The max depth of the network to be activated
	if err != nil {
		neat.WarnLog(
			fmt.Sprintf("Failed to estimate maximal depth of the network with loop:\n%s\nUsing default dpeth: %d",
				org.Genotype, net_depth))
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
	org.Phenotype.LoadSensors(inputs)

	// propagate input through the phenotype net

	// Relax phenotype net and get output
	_, err = org.Phenotype.Activate()
	if err != nil {
		neat.ErrorLog("Failed to activate network")
		return nil, err
	}

	// use depth to ensure relaxation at each layer
	for relax := 0; relax <= net_depth; relax++ {
		_, err = org.Phenotype.Activate()
		if err != nil {
			neat.ErrorLog("Failed to activate network")
			return nil, err
		}
	}

	return &env
}

// To execute a time step of the maze simulation evaluation within given Environment for provided Organism
// Returns fitness value of evaluated organism
func mazeSimulationStep(env *Environment, org *genetics.Organism) (float64, error) {
	// get simulation parameters as inputs to organism's network
	inputs, err := env.GetInputs()
	if err != nil {
		return -1.0, err
	}
	org.Phenotype.LoadSensors(inputs)
	_, err = org.Phenotype.Activate()
	if err != nil {
		neat.ErrorLog("Failed to activate network")
		return -1.0, err
	}

	// use the net's outputs to change heading and velocity of maze agent
	err = env.ApplyOutputs(org.Phenotype.Outputs[0].Activation, org.Phenotype.Outputs[1].Activation)
	if err != nil {
		neat.ErrorLog("Failed to apply outputs")
		return -1.0, err
	}

	// update the environment
	err = env.Update()
	if err != nil {
		neat.ErrorLog("Failed to update environment")
		return -1.0, err
	}

	dist, err := env.distanceToExit()
	if err != nil {
		neat.ErrorLog("Failed to estimate distance to maze exit")
		return -1.0, err
	}
	if dist < 1 {
		dist = 1
	}

	fitness := 5.0 / dist

	return fitness
}