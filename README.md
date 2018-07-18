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

* [Novelty Search and the Problem with Objectives][4]
* [EVOLUTION THROUGH THE SEARCH FOR NOVELTY][5]

#### System Requirements

The source code written and compiled against GO 1.9.x.

## Installation

Make sure that you have at least GO 1.9.x. environment installed onto your system and execute following command:
```bash

go get -t github.com/yaricom/goNEAT_NS
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

![alt text][seed_genome_graph]

Thus the seed genome of maze solving agent has following configuration (see [seed genome](data/mazestartgenes) for details):

* ten input (sensor) neurons: six for range finders [-90.0, -45.0, 0.0, 45.0, 90.0, -180.0] plus four for slice radar sensors with 45 degree FOV [RIGHT, UP, LEFT, DOWN] (blue)
* two output neurons: linear and angular velocity controlling effectors (red)
* one hidden neuron to introduce non linearity (green)
* one bias neuron to avoid zero saturation when input neurons is not activated (yellow)

During NEAT algorithm execution with Novelty Search optimization the provided seed genome will be complexified by
adding new nodes/links and adjusting link weights.

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
go run executor.go -out ./out/medium_mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/medium_maze.txt -experiment MAZENS

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/medium_maze.txt is a maze environment configuration.

This command will execute one trial with 2000 generations (or less if winner is found) over population of 250 organisms.

The experiment results will be similar to the following:

```
Average
	Winner Nodes:	16.0
	Winner Genes:	29.0
	Winner Evals:	16111.0
Mean
	Complexity:	28.5
	Diversity:	22.1
	Age:		64.8
```

Where:
- **Winner nodes/genes** is number of units and links between in produced Neural Network which was able to solve XOR problem.
- **Winner evals** is the number of evaluations of intermediate organisms/genomes before winner was found.
- **Mean Complexity** is an average complexity (number of nodes + number of links) of best organisms per epoch for all epochs.
- **Mean Diversity** is an average diversity (number of species) per epoch for all epochs
- **Mean Age** is an average age of surviving species per epoch for all epochs

![alt text][mazens_medium_winner_genome_graph]

After 64 generations was found near optimal winner genome configuration able to control maze solving agent. The artificial
neural network produced by this genome has only 16 units (neurons) with three hidden neurons.

During the experiment novelty search optimization resulted in growing two additional hidden units (neurons) and
introducing recurrent link at one of the output neurons (#13). The recurrent link at the output neuron seems to have
extreme importance as it's introduced at each winner genome configurations generated by solution.

Introduced genome was able to solve maze and find exit with spatial error about 1.9% at the exit point.

![alt text][mazens_medium_winner_records]

Above is a rendering of maze solving simulation by agents controlled by ANNs which is generated from genomes of all organisms
introduced into population until winner is found. The agents is *color coded* depending on which species the source organism
belongs. The fitness of agent is measured as a relative distance between it's final destination and maze exit after running simulation
for particular number of time steps (400 in our setup).

The initial agent position is at the top-left corner marked with green circle and maze exit at the bottom-right marked with red circle.

The top plot shows final destinations of the most fit agents (fitness >= 0.8) and bottom is the rest. The results is given
for experimental run with winner genome configuration presented above. At that experiment was produced 32 species among which
the most fit ones has amounted to eight.

#### To run experiment with hard difficulty maze map execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/hard_mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/hard_maze.txt -experiment MAZENS

```

![alt text][mazens_hard_winner_genome_graph]

After 109 generations of population was found near optimal winner genome configuration able to guide maze solving agent through
hard maze. The artificial neural network produced by this genome has only 17 units (neurons) with four hidden neurons.

The optimal genome configuration produced by growing three additional hidden units and multiple new links compared to seed
genome. It's interesting to note that recurrent link at output neuron #13 (linear velocity effector) was routed through
two hidden neurons in contrast with medium maze where neuron #13 was simply linked to itself. This may result in more complex
behaviour learned especially taking into account that link pass through neuron #42 (affected by range finder: UP; radar: LEFT)
and neuron #12 which is introduced at the seed genome and serves as *main relay control unit* relaying input signals from
all input (sensor) neurons to all output (effector) neurons.

It worth to mention about special role of mentioned neuron #42 which seems to encode complex learned behaviour about general
direction to the maze exit point.

Other important point to note is about possible learned behaviour encoded by hidden neuron #297 - it's affected by input
range finder sensors detecting distance to obstacles at *DOWN* and *DOWN-RIGHT* direction. Looking at maze configuration it
can be seen that most trapping cul-de-sacs can be escaped by moving exactly in this direction. So, we can assume that this
neuron participates in strategy to avoid vertical traps with local optimum fitness values based on distance to the exit.

![alt text][mazens_hard_winner_records]


## Auxiliary Tools

During this project development was created several tools to help with results visualization and pre-/post-processing.

### Genome to GraphML converter

Helps with conversion of genome data into GraphML to render genome as a graph with help of specialized software such as
(Cytoscape)[http://www.cytoscape.org]

Use following command to run it:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
python tools/genome_utils.py [in_file] --out [out_file]

```
Where:

- **in_file** the input file to read genome data from, e.g [seed genome](data/mazestartgenes)
- **out_file** the output file to write GraphML presentation

### The agents' data records visualizer for maze solving simulations

Allows to visualize recorded data of maze solving agents color coded by species they belong and separated into two groups:
the best and other.

Use following command to run it:

```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run tools/maze_utils.go -records [records_file] -maze [maze_file] -out [out_file] -width [width] -height [height]

```
Where:

- **records_file** the file holding recorded data of maze solving by population agents
- **maze_file** the maze configuration file, e.g. [medium_maze.txt](data/medium_maze.txt)
- **out_file** the output file [PNG]
- **width** the plot canvas width
- **height** the plot canvas height


## Credits

The original C++ NEAT implementation created by Kenneth Stanley, see: [NEAT][1]

This source code maintained and managed by Iaroslav Omelianenko

Other NEAT implementations may be found at [NEAT Software Catalog][2]

[1]:http://www.cs.ucf.edu/~kstanley/neat.html
[2]:http://eplex.cs.ucf.edu/neat_software/
[3]:https://github.com/yaricom/goNEAT
[4]:http://eplex.cs.ucf.edu/papers/lehman_gptp11.pdf
[5]:http://joellehman.com/lehman-dissertation.pdf


[seed_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/seed_genome.png "The seed genome graph"
[mazens_medium_winner_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_medium_16/mazens_winner_16.png "The graph for near optimal winner genome generated by novelty search for medium maze"
[mazens_medium_winner_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_medium_16/ns_medium_maze_16.png "The plot of maze agent records for medium maze"
[mazens_hard_winner_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_hard_17/17_hard_mazens_winner.png "The graph for near optimal winner genome generated by novelty search for hard maze"
[mazens_hard_winner_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_hard_17/17_ns_hard_maze.png "The plot of maze agent records for hard maze"