## Overview

This repository provides implementation of [Neuro-Evolution of Augmented Topologies (NEAT)][1] with Novelty Search
optimization implemented in GoLang.

The Neuro-Evolution (NE) is an artificial evolution of Neural Networks (NN) using genetic algorithms in order to find
optimal NN parameters and topology. Neuro-Evolution of NN may assume search for optimal weights of connections between
NN nodes as well as search for optimal topology of resulting NN. The NEAT method implemented in this work do search for
both: optimal connections weights and topology for given task (number of NN nodes per layer and their interconnections).

The Novelty Search optimization allows to solve deceptive tasks with strong local optima which can not be solved by
traditional objective-based fitness optimization functions. One of such problems is maze navigation where non-objective
search methods like novelty search may outperform more traditional objective-based search methods. Our goal in this
research is to test this hypothesis. For more information about Novelty Search optimization please refer to:

* [Novelty Search and the Problem with Objectives](http://eplex.cs.ucf.edu/papers/lehman_gptp11.pdf)
* [EVOLUTION THROUGH THE SEARCH FOR NOVELTY](http://joellehman.com/lehman-dissertation.pdf)

#### System Requirements

The source code written and compiled against GO 1.9.x.

## Installation

Make sure that you have at least GO 1.9.x. environment installed onto your system and execute following command:
```bash

go get github.com/yaricom/goNEAT_NS
```

This project is dependent on [goNEAT][3] project which will be installed automatically when command above executed.

## Experiment Overview

An illustrative domain for testing novelty search should have a deceptive fitness landscape. In such a domain, a search
algorithm following the fitness gradient may perform worse than an algorithm following novelty gradients because novelty
cannot be deceived with respect to the objective; it ignores objective fitness entirely.

A compelling, easily-visualized domain with this property is a two-dimensional maze navigation task, wherein an agent
must navigate through a maze to a chosen goal point. A reasonable fitness function for such a domain is how close the
maze navigator is to the goal at the end of the evaluation. Thus, dead ends that lead close to the goal are local optima
to which an objective-based algorithm may converge, which makes a good model for deceptive problems in general. Moreover,
by varying the structure of the maze and the starting and goal point of the robot, various classes of problems can
be modeled (e.g. removing the exterior walls of a maze results in a more unconstrained problem).

### The Maze Agent

A maze agent controlling by Artificial Neural Network [ANN] must navigate from starting point to the exit within given number
of time steps, i.e. in fixed time. This excludes dumb random search covering majority of maze locations which will take
great amount of time steps to be executed. The task is complicated by cul-de-sacs that prevent a direct route and that
create local optima in the fitness landscape.

The agent has *six rangefinders* that indicate the distance to the nearest obstacle and *four pie-slice radar sensors* that
fire when the goal is within the pie-slice. The agent’s two effectors result in forces that respectively turn and propel
the robot, i.e. change it's linear and angular velocity.

As a result the seed genome of organism describing maze agent's behaviour has ten input nodes/neurons, two output nodes,
one bias node and one hidden unit to model non linearity (see [seed genome](data/mazestartgenes) for details).
During NEAT algorithm execution with Novelty Search optimization provided seed genome will become more complex by
adding new nodes/links and adjusting link weights.

![alt text][seed_genome_graph]

## Experiments and Performance evaluation

In order to test hypothesis that novelty search based optimization outperforms traditional objective-based
optimization two experiments will be studied:

* the maze navigation with novelty search optimization
* the maze navigation with objective-based optimization

We will execute experiments within maze configurations of two difficulty levels:

* medium difficulty map
* hard difficulty map

### 1. The Maze Navigation with Novelty Search optimization

In this experiment evaluated the performance of maze agent controlled by ANN which is created by NEAT algorithm with
Novelty Search based optimization. The mentioned optimization is based on *novelty metric* calculation for each agent
after particular time steps of maze navigation simulation for that agent is performed. The novelty metric biases the
search in a fundamentally different way than the objective-based fitness function based on distance from agent to exit.
The novelty metric determines the behavior-space through which search will proceed. Therefore, because what is important
in a maze is where the traverser ends, for the maze domain, the behavior of a navigator is defined as its ending position.
The novelty metric is then the squared Euclidean distance between the ending positions of two individuals.


The effect of this novelty metric is to reward the robot for ending in a place where none have ended before; the method
of traversal is ignored. This measure reflects that what is important is reaching a certain location (i.e. the goal)
rather than the method of locomotion. Thus, although the novelty metric has no knowledge of the final goal, a solution
that reaches the goal can appear novel. In addition, the comparison between fitness-based and novelty-based search is
fair because both scores are computed only based on the distance of the final position of the robot from other points.

#### To run experiment with medium difficulty maze map execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/medium_maze.txt -experiment MAZENS

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/medium_maze.txt is a maze environment configuration.

This command will execute one trial with 2000 generations (or less if winner is found) over population of 250 organisms.

The experiment results will be similar to the following:

```
Average
	Winner Nodes:	17.0
	Winner Genes:	30.0
	Winner Evals:	70041.0
Mean
	Complexity:	45.6
	Diversity:	27.4
	Age:		215.8
```

Where:
- **Winner nodes/genes** is number of units and links between in produced Neural Network which was able to solve XOR problem.
- **Winner evals** is the number of evaluations of intermediate organisms/genomes before winner was found.
- **Mean Complexity** is an average complexity (number of nodes + number of links) of best organisms per epoch for all epochs.
- **Mean Diversity** is an average diversity (number of species) per epoch for all epochs
- **Mean Age** is an average age of surviving species per epoch for all epochs

![alt text][mazens_winner_genome_graph]

After 281 generations was found near optimal winner genome configuration able to control maze solving agent. The artificial
neural network produced by this genome has only 17 units (neurons) with three hidden neurons.

During the experiment novelty search optimization resulted in growing three additional hidden units (neurons) and
introducing recurrent link at one of the output neurons (#13). The recurrent link at the output neuron seems to have
extreme importance as it's introduced at each winner genome configurations generated by solution.

Introduced genome was able to solve maze and find exit with spatial error about 0.8% at the exit point.

#### To run experiment with hard difficulty maze map execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/hard_maze.txt -experiment MAZENS

```



## Credits

The original C++ NEAT implementation created by Kenneth Stanley, see: [NEAT][1]

This source code maintained and managed by Iaroslav Omelianenko

Other NEAT implementations may be found at [NEAT Software Catalog][2]

[1]:http://www.cs.ucf.edu/~kstanley/neat.html
[2]:http://eplex.cs.ucf.edu/neat_software/
[3]:https://github.com/yaricom/goNEAT


[seed_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/seed_genome.png "The seed genome graph"
[mazens_winner_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/mazens_winner_genome.png "The graph for near optimal novelty search generated winner genome"
