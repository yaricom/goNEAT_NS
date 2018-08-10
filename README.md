[![Build Status](https://travis-ci.org/yaricom/goNEAT_NS.svg?branch=master)](https://travis-ci.org/yaricom/goNEAT_NS) [![GoDoc](https://godoc.org/github.com/yaricom/goNEAT_NS/neatns?status.svg)](https://godoc.org/github.com/yaricom/goNEAT_NS/neatns)

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
research is to test this hypothesis.

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

The agent has *six rangefinders* that indicate the distance to the nearest obstacle and *four pie-slice radar sensors* that act
as a compass towards the goal (maze exit), activating when a line from the goal to the center of the agent falls within the pie-slice.

The radar sensors has following FOV zones relative to agent's heading direction: FRONT, LEFT, BACK, RIGHT, or in degrees relative to the
agent's heading: (315.0 ~ 405.0), (45.0 ~ 135.0), (135.0 ~ 225.0), (225.0 ~ 315.0). The range finder sensors
monitor following directions relative to the agent heading: RIGHT, FRONT-RIGHT, FRONT, FRONT-LEFT, LEFT, BACK or in degrees relative
to agent's heading: -90.0, -45.0, 0.0, 45.0, 90.0, -180.0.

The agent’s two effectors result in forces that respectively turn and propel the robot, i.e. change it's *linear and angular velocity*.

![alt text][seed_genome_graph]

Thus the seed genome of maze solving agent has following configuration (see [seed genome](data/mazestartgenes) for details):

* ten input (sensor) neurons: six for range finders [RIGHT, FRONT-RIGHT, FRONT, FRONT-LEFT, LEFT, BACK] plus four for slice
radar sensors with 45 degree FOV [FRONT, LEFT, BACK, RIGHT] (blue)
* two output neurons: angular (neuron #13) and linear (neuron #14) velocity controlling effectors (red)
* one hidden neuron (#12) to introduce non linearity (green)
* one bias neuron (#1) to avoid zero saturation when input neurons is not activated (yellow)

The input neurons as following:

* Range Finders: #2 - RIGHT, #3 - FRONT-RIGHT, #4 - FRONT, #5 - FRONT-LEFT, #6 - LEFT, #7 - BACK
* Radar Sensors: #8 - FRONT, #9 - LEFT, #10 - BACK, #11 - RIGHT


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

### 1. The Maze Navigation with Novelty Search Optimization

In this experiment evaluated the performance of maze agent controlled by ANN which is created by NEAT algorithm with
Novelty Search based optimization. The mentioned optimization is based on *novelty metric* calculation for each agent
after particular time steps of maze navigation simulation for that agent is performed. The novelty metric biases the
search in a fundamentally different way than the objective-based fitness function based on distance from agent to exit.
The novelty metric determines the behavior-space through which search will proceed. Therefore, because what is important
in a maze is where the solving agent ends, for the maze domain, the behavior of a navigator is defined as its ending position.
The novelty metric is then the N-nearest neighbor distance novelty between the ending positions of all known solving agents.


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

After 64 generations was found near optimal winner genome configuration able to control maze solving agent and find the exit
with spatial error about 1.9% at the exit point. The artificial neural network produced by this genome has only 16 units
(neurons) with three hidden neurons.

![alt text][mazens_medium_winner_genome_graph]

During the experiment novelty search optimization resulted in growing two additional hidden units (neurons) and
introducing recurrent link at the output neuron #13 (angular velocity effector). The recurrent link at that output neuron
seems to have extreme importance as it's introduced at each winner genome configuration generated by solution. It seems
reasonable because neuron #13 effects steering of the agent and need to learn more complex behaviour than neuron #14 (linear
velocity control).

It's interesting to note hidden neuron #91 which seems to learn complex behaviour of steering to the maze exit
when it is detected rightward or behind of the agent. We've made such assumptions because of its connections with input
sensors #2, #7 (range finders: *RIGHT, BACK*) and #10, #11 (radar sensors: *BACK, RIGHT*).

The hidden neuron #293 connected with input sensor #11 (radar sensor: *RIGHT*) learned to affect agent's steering in the direction
of maze exit as most of the times it is at the right bottom relative to the agent.

The hidden neuron #12 which is introduced in seed genome operates as main control-and-relay switch relaying signals from sensors
 and other hidden neurons to the effectors (neurons #13, #14).

![alt text][mazens_medium_winner_records]

Above is a rendering of the maze solving simulation by agents controlled with ANNs generated from genomes of all organisms
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
hard maze and approach the maze exit with spatial error of 2.5%. The artificial neural network produced by this genome
has only 17 units (neurons) with four hidden neurons to model complex learned behaviour.

The optimal genome configuration produced by growing three additional hidden units and multiple new links compared to seed
genome. It's interesting to note that recurrent link at output neuron #13 (angular velocity effector) was routed through
two hidden neurons in contrast with medium maze where neuron #13 was simply linked to itself. This may result in more complex
behaviour learned especially taking into account that link pass through neuron #42 affected by range finder: *LEFT* and radar: *BACK*.
The neuron #42 also affected by connection with neuron #643 (affected by range finder: *LEFT*). As a result we may assume that
it learned how steer agent when maze exit is behind and wall is at the left of it, i.e. to follow the left wall by moving
forward.

Other important point to note is about possible learned behaviour encoded by hidden neuron #297 - it's affected by input
range finder sensors detecting distance to obstacles at *RIGHT* and *FRONT* direction. Looking at maze configuration we
may assume that this neuron learned to avoid left chamber trap with extremely strong local optimum of fitness based on
the distance to the maze exit.

![alt text][mazens_hard_winner_records]

Above is the visualization of hard maze solving by all agents until winner is found. The initial agent position is at the
bottom-left and maze exit at the top-left of the maze. The agents is color coded based on species they belong. So, each
dot of similar color is the final position of agent controlled by organism belonging to the same species.

The top plot shows final destinations of the most fit agents (fitness >= 0.8) and bottom is the rest. The fitness of agent
is measured as distance from it's final position to the maze exit after 400 time steps of simulation.

From the plot we can see that winner species produced organisms that control agents in such a way that its final destinations
is evenly distributed through the maze. As a result it was possible to produce control ANN able to solve the maze.


### 2. The Maze Navigation with Objective-Based Fitness Optimization

In this experiment evaluated the performance of maze agent controlled by ANN which is created by NEAT algorithm with
*objective-based fitness* optimization. The mentioned optimization is based on maximizing solving agent's fitness by following
its objective, i.e. the distance from agent to exit. As with previous experiment the behavior of a navigator is defined as its
ending position in a maze. The fitness function is then the squared Euclidean distance between the ending position of the agent
and maze exit.

The effect of this fitness function is to reward the solving agent for ending in a place as close to the maze exit as possible.

#### To run experiment with medium difficulty maze map execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazeobj -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/medium_maze.txt -experiment MAZEOBJ

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/medium_maze.txt is a maze environment configuration.

This command will execute one trial with 2000 generations (or less if winner is found) over population of 250 organisms.

The experiment results will be similar to the following:

```
Average
	Winner Nodes:	22.0
	Winner Genes:	49.0
	Winner Evals:	62168.0
Mean
	Complexity:	44.6
	Diversity:	27.5
	Age:		222.0
```

![alt text][mazeobj_medium_winner_genome_graph]

After 248 generations of population was found near optimal winner genome configuration able to guide maze solving agent through
medium maze and approach the maze exit with spatial error of 1.8%. The artificial neural network produced by this genome
has 22 units (neurons) with nine hidden neurons to model complex learned behaviour.

The genotype of the winning agent presented above has more complicated structure compared to the near optimal genome created
by *Novelty Search* based optimization from the first experiment with more redundant neurons and links. Due to added
complexity, the produced organism is less energy efficient and harder to execute at the inference time.


![alt text][mazeobj_medium_winner_records]

Above is a rendering of the maze solving simulation by agents controlled with ANNs generated from genomes of all organisms
introduced into population until winner is found. The agents is color coded depending on which species the source organism
belongs. The fitness of agent is measured as a relative distance between it's final destination and maze exit after running
simulation for particular number of time steps (400 in our setup).

The initial agent position is at the top-left corner marked with green circle and maze exit at the bottom-right marked with red circle.

By comparing it with simulation based on Novelty Search optimization it may be seen that agent's final destinations is less
evenly distributed through the maze space and some areas left completely unexplored.

#### To run experiment with hard difficulty maze map execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazeobj_hard -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/hard_maze.txt -experiment MAZEOBJ

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/hard_maze.txt is a maze environment configuration.

After 10 trials objective-based optimization function was **unable to produce any successful hard maze solving agent**. The hard
maze has deceptive cul-de-sacs, where objective-based fitness function has strong local optima. Thus when neuroevolution is
based on this fitness function it's unable to make a leap to the next level of genome complexity able to solve hard maze
deceptive traps. At the same time **novelty search based fitness function** was able to produce maze solvers able to crack
hard maze configuration with the same ease as medium maze configuration.

![alt text][mazeobj_hard_failure_records]

Above is renderings of hard maze solving agents' final destinations for several failed trials. For all 10 executed trials
the renderings looks similar and by examining them it can be easy detected mentioned local optima traps which prevent any
produced organism from solving hard maze configuration.


## Discussion

In this work we have tested two approaches to perform fitness function optimization with NEAT algorithm: novelty search and
objective-based. The novelty search optimization was found as outperforming method for solving of deceptive tasks when strong local
optima present, such as maze solving. Our experiments based on two maze environments configurations: medium and hard maze.

### Medium maze results

With medium maze configuration both fitness function optimization methods was able to produce agents able to solve the maze:

* the Novelty Search based agent was able to solve maze in ten from ten trials
* the Objective-Based agent was able to solve medium maze in nine from ten trials

The novelty search optimization also resulted in producing more energy efficient and elegant genome for solver agent. The
absolute winner with NS optimization has only 15 neurons with 19 links between (Fitness: 0.984) compared to objective-based
optimization where best agent has 60 neurons with 214 links between (Fitness: 0.987). The provided fitness value describe
how close final agent's position to the maze exit after 400 time steps (where 1.0 means exact match). Full statistics of
experiment provided further.

```
Medium Maze Objective-Based:
============================
+++ Solved 9 trials from 10 +++

Champion found in 3 trial run
	Winner Nodes:	60
	Winner Genes:	214
	Winner Evals:	235020

	Diversity:	26
	Complexity:	274
	Age:		320
	Fitness:	1.0

Average among winners
	Winner Nodes:	34.0
	Winner Genes:	110.2
	Winner Evals:	160756.8

	Diversity:	20.1
	Complexity:	144.2
	Age:		251.0
	Fitness:	1.0

Averages for all organisms evaluated during experiment
	Diversity:	18.5
	Complexity:	105.5
	Age:		165.5
	Fitness:	0.8

Medium Maze Novelty Search:
============================
+++ Solved 10 trials from 10 +++

Champion found in 3 trial run
	Winner Nodes:	15
	Winner Genes:	19
	Winner Evals:	13838

	Diversity:	23
	Complexity:	34
	Age:		45
	Fitness:	1.0

Average among winners
	Winner Nodes:	17.1
	Winner Genes:	29.1
	Winner Evals:	25132.1

	Diversity:	23.0
	Complexity:	46.2
	Age:		78.2
	Fitness:	1.0

Averages for all organisms evaluated during experiment
	Diversity:	17.6
	Complexity:	43.5
	Age:		41.1
	Fitness:	0.5
```


### Hard maze results

With hard maze configuration objective-based optimization method *failed to produce any agent able to solve this maze.*
At the same time Novelty Search based optimization able to avoid deceptive strong local optima introduced in hard maze
and produce effective solver agents in less than 300 generations over the same ten trial executions.

```
Hard Maze Novelty Search:
============================
+++ Solved 10 trials from 10 +++

Champion found in 4 trial run
	Winner Nodes:	27
	Winner Genes:	66
	Winner Evals:	64314

	Diversity:	27
	Complexity:	93
	Age:		235
	Fitness:	1.0

Average among winners
	Winner Nodes:	17.6
	Winner Genes:	34.2
	Winner Evals:	29530.9

	Diversity:	25.6
	Complexity:	51.8
	Age:		75.5
	Fitness:	1.0

Averages for all organisms evaluated during experiment
	Diversity:	18.6
	Complexity:	46.4
	Age:		42.7
	Fitness:	0.3
```

### Conclusion

As it was shown by experimental data, the Novelty Search optimization, where fitness of agent is based on novelty of the solution
it was able to find, considerably outperforms traditional objective-based optimization and even was able to solve task where
traditional method failed completely.

We believe that novelty search optimization can be successfully applied to produce optimal solving agents in many areas where
 strong deceptive local fitness optima is blocking traditional objective-based methods from finding optimal or any solutions.

For more information about Novelty Search optimization please refer to original works:

* [Novelty Search and the Problem with Objectives][4]
* [EVOLUTION THROUGH THE SEARCH FOR NOVELTY][5]

## Auxiliary Tools

During this project development was created several tools to help with results visualization and pre-/post-processing.

### Genome to GraphML converter

Helps with conversion of genome data into GraphML to render genome as a graph with help of specialized software such as
[Cytoscape](http://www.cytoscape.org)

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
[mazens_medium_winner_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_medium_16/ns_medium_maze_16.png "The plot of maze agent records for medium maze when novelty search optimization applied"

[mazens_hard_winner_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_hard_17/17_hard_mazens_winner.png "The graph for near optimal winner genome generated by novelty search for hard maze"
[mazens_hard_winner_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/NS_hard_17/17_ns_hard_maze.png "The plot of maze agent records for hard maze when novelty search optimization applied"

[mazeobj_medium_winner_genome_graph]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/OBJ_medium_22/22_maze_obj_winner.png "The graph for near optimal winner genome generated by objective-based fitness optimization for medium maze"
[mazeobj_medium_winner_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/OBJ_medium_22/22_obj_medium_maze.png "The plot of maze agent records for medium maze when objective-based fitness optimization applied"

[mazeobj_hard_failure_records]: https://github.com/yaricom/goNEAT_NS/blob/master/contents/OBJ_hard/hard_mazeobj.png "The plot with hard maze solving agents final positions for hard maze when objective-based fitness optimization applied"