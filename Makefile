#
# Go parameters
#
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test -count=1
GOGET = $(GOCMD) get
GORUN = $(GOCMD) run

# The common parameters
BINARY_NAME = goneatns
OUT_DIR = out

# The default parameters to run the experiment
DATA_DIR=./data
TRIALS_NUMBER=10
LOG_LEVEL=info


# The default targets to run
#
all: test

# The target to run Maze Novelty Search Experiment with medium Maze
#
run-maze-ns-medium:
	$(GORUN) executor.go -out $(OUT_DIR)/mazens \
							-context $(DATA_DIR)/maze.neat \
							-genome $(DATA_DIR)/mazestartgenes.yml \
							-maze $(DATA_DIR)/medium_maze.txt \
							-experiment MAZENS \
							-trials $(TRIALS_NUMBER) \
							-log_level $(LOG_LEVEL)

# The target to run Maze Novelty Search Experiment hard Maze
#
run-maze-ns-hard:
	$(GORUN) executor.go -out $(OUT_DIR)/mazens \
							-context $(DATA_DIR)/maze.neat \
							-genome $(DATA_DIR)/mazestartgenes.yml \
							-maze $(DATA_DIR)/hard_maze.txt \
							-experiment MAZENS \
							-trials $(TRIALS_NUMBER) \
							-log_level $(LOG_LEVEL)

# The target to run Maze Objective Search Experiment with medium Maze
#
run-maze-objective-medium:
	$(GORUN) executor.go -out $(OUT_DIR)/mazens \
							-context $(DATA_DIR)/maze.neat \
							-genome $(DATA_DIR)/mazestartgenes.yml \
							-maze $(DATA_DIR)/medium_maze.txt \
							-experiment MAZEOBJ \
							-trials $(TRIALS_NUMBER) \
							-log_level $(LOG_LEVEL)

# The target to run Maze Novelty Search Experiment hard Maze
#
run-maze-objective-hard:
	$(GORUN) executor.go -out $(OUT_DIR)/mazens \
							-context $(DATA_DIR)/maze.neat \
							-genome $(DATA_DIR)/mazestartgenes.yml \
							-maze $(DATA_DIR)/hard_maze.txt \
							-experiment MAZEOBJ \
							-trials $(TRIALS_NUMBER) \
							-log_level $(LOG_LEVEL)

# Run unit tests in short mode
#
test-short:
	$(GOTEST) -v --short ./...

# Run all unit tests
#
test:
	$(GOTEST) -v ./...

# Builds binary
#
build: | $(OUT_DIR)
	$(GOBUILD) -o $(OUT_DIR)/$(BINARY_NAME) -v

# Creates the output directory for build artefacts
#
$(OUT_DIR):
	mkdir -p $@

#
# Clean build targets
#
clean:
	$(GOCLEAN)
	rm -f $(OUT_DIR)/$(BINARY_NAME)