# goNEAT_NS
The NeuroEvolution of Augmenting Topologies (NEAT) with Novelty Search implementation in GoLang

#### System Requirements
The source code written and compiled against GO 1.9.x.

## Installation
Make sure that you have at least GO 1.9.x. environment installed onto your system and execute following command:
```bash

go get github.com/yaricom/goNEAT_NS
```

## 1. The Maze NS Experiments

To run this experiment execute following commands:
```bash

cd $GOPATH/src/github.com/yaricom/goNEAT_NS
go run executor.go -out ./out/mazens -context ./data/maze.neat -genome ./data/mazestartgenes -maze ./data/medium_maze.txt -experiment MAZENS

```
Where: ./data/maze.neat is the configuration of NEAT execution context, ./data/mazestartgenes is the start genome
configuration, and ./data/medium_maze.txt is a maze environment configuration.
