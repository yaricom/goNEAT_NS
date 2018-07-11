## Overview
This repository provides implementation of [NeuroEvolution of Augmenting Topologies (NEAT)][1] with Novelty Search
optimization implemented in GoLang.

The Neuroevolution (NE) is an artificial evolution of Neural Networks (NN) using genetic algorithms in order to find
optimal NN parameters and topology. Neuroevolution of NN may assume search for optimal weights of connections between
NN nodes as well as search for optimal topology of resulting NN. The NEAT method implemented in this work do search for
both: optimal connections weights and topology for given task (number of NN nodes per layer and their interconnections).

The Novelty Search optimization allows to solve deceptive tasks with strong local optimums which can not be solved by
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

### The maze agent
A maze agent controlling by Artificial Neural Network must navigate from starting point to the exit within given number
of time steps, i.e. in fixed time. This excludes dumb random search covering majority of maze locations which will take
great amount of time steps to be executed. The task is complicated by cul-de-sacs that prevent a direct route and that
create local optima in the fitness landscape.

The agent has six rangefinders that indicate the distance to the nearest obstacle and four pie-slice radar sensors that
fire when the goal is within the pie-slice. The agent’s two effectors result in forces that respectively turn and propel
the robot, i.e. change it's linear and angular velocity.

As a result the seed genome of organism describing maze agent's behaviour has ten input nodes/neurons, two output nodes,
one bias node and one hidden unit to model non linearity. During NEAT algorithm execution with Novelty Search optimization
provided seed genome will become more complex by adding new nodes/links and adjusting link weights.

### Performance evaluation
In order to test hypothesis that novelty search based optimization outperforms traditional fitness objective-based
optimization two experiments will be studied:

* the maze navigation with novelty search optimization
* the maze navigation with fitness objective-based optimization

The experiments results follow below.

### 1. The Maze NS Experiments

To run this experiment execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/medium_maze.txt -experiment MAZENS

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/medium_maze.txt is a maze environment configuration.


## Credits

The original C++ NEAT implementation created by Kenneth Stanley, see: [NEAT][1]

This source code maintained and managed by Iaroslav Omelianenko

Other NEAT implementations may be found at [NEAT Software Catalog][2]

[1]:http://www.cs.ucf.edu/~kstanley/neat.html
[2]:http://eplex.cs.ucf.edu/neat_software/
[3]:https://github.com/yaricom/goNEAT
