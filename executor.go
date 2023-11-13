package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT_NS/v4/examples/maze"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// The experiment runner boilerplate code
func main() {
	var outDirPath = flag.String("out", "./out", "The output directory to store results.")
	var contextPath = flag.String("context", "./data/maze.neat", "The execution context configuration file.")
	var genomePath = flag.String("genome", "./data/mazestartgenes", "The seed genome to start with.")
	var mazeConfigPath = flag.String("maze", "./data/medium_maze.txt", "The maze environment configuration file.")
	var experimentName = flag.String("experiment", "MAZENS", "The name of experiment to run. [MAZENS, MAZEOBJ]")
	var timeSteps = flag.Int("timesteps", 400, "The number of time steps for maze simulation per organism.")
	var timeStepsSample = flag.Int("timesteps_sample", 1000, "The sample size to store agent path when doing maze simulation.")
	var speciesTarget = flag.Int("species_target", 20, "The target number of species to maintain.")
	var speciesCompatAdjustFreq = flag.Int("species_adjust_freq", 10, "The frequency of species compatibility theshold adjustments when trying to maintain their number.")
	var trialsCount = flag.Int("trials", 0, "The number of trials for experiment. Overrides the one set in configuration.")
	var logLevel = flag.String("log_level", "", "The logger level to be used. Overrides the one set in configuration.")
	var exitRange = flag.Float64("exit_range", 5.0, "The range around maze exit point to test if agent coordinates is within to be considered as solved successfully")

	flag.Parse()

	// Seed the random-number generator with current time so that
	// the numbers will be different every time we run.
	seed := time.Now().Unix()
	rand.Seed(seed)

	// Load NEAT options
	neatOptions, err := neat.ReadNeatOptionsFromFile(*contextPath)
	if err != nil {
		log.Fatal("Failed to load NEAT options: ", err)
	}

	// Load Genome
	log.Printf("Loading start genome for %s experiment from file '%s'\n", *experimentName, *genomePath)
	reader, err := genetics.NewGenomeReaderFromFile(*genomePath)
	if err != nil {
		log.Fatalf("Failed to open genome file, reason: '%s'", err)
	}
	startGenome, err := reader.Read()
	if err != nil {
		log.Fatalf("Failed to read start genome, reason: '%s'", err)
	}
	fmt.Println(startGenome)

	// Load maze environment
	log.Printf("Reading maze environment: %s\n", *mazeConfigPath)
	mazeFile, err := os.Open(*mazeConfigPath)
	var environment *maze.Environment
	if err == nil {
		environment, err = maze.ReadEnvironment(mazeFile)
		if environment != nil {
			environment.TimeSteps = *timeSteps
			environment.SampleSize = *timeStepsSample
			environment.ExitFoundRange = *exitRange
		}
		log.Println(environment)
	}
	if err != nil {
		log.Fatal("Failed to read maze environment configuration: ", err)
	}

	// Check if output dir exists
	outDir := *outDirPath
	if _, err = os.Stat(outDir); err == nil {
		// backup it
		backUpDir := fmt.Sprintf("%s-%s", outDir, time.Now().Format("2006-01-02T15_04_05"))
		// clear it
		err = os.Rename(outDir, backUpDir)
		if err != nil {
			log.Fatal("Failed to do previous results backup: ", err)
		}
	}
	// create output dir
	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create output directory: ", err)
	}

	// Override context configuration parameters with ones set from command line
	if *trialsCount > 0 {
		neatOptions.NumRuns = *trialsCount
	}
	if len(*logLevel) > 0 {
		if err = neat.InitLogger(*logLevel); err != nil {
			log.Fatal("Failed to initialize logger: ", err)
		}
	}

	// create experiment
	expt := experiment.Experiment{
		Id:       0,
		Trials:   make(experiment.Trials, neatOptions.NumRuns),
		RandSeed: seed,
	}
	var generationEvaluator experiment.GenerationEvaluator
	var trialObserver experiment.TrialRunObserver
	if *experimentName == "MAZENS" {
		generationEvaluator, trialObserver = maze.NewNoveltySearchEvaluator(
			outDir, environment, *speciesTarget, *speciesCompatAdjustFreq)
	} else if *experimentName == "MAZEOBJ" {
		generationEvaluator, trialObserver = maze.NewMazeObjectiveEvaluator(
			outDir, environment, *speciesTarget, *speciesCompatAdjustFreq)
	} else {
		log.Fatalf("Unsupported experiment name requested: %s\n", *experimentName)
	}

	// prepare to execute
	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	// run experiment in the separate GO routine
	go func() {
		if err = expt.Execute(neat.NewContext(ctx, neatOptions), startGenome, generationEvaluator, trialObserver); err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	// register handler to wait for termination signals
	//
	go func(cancel context.CancelFunc) {
		fmt.Println("\nPress Ctrl+C to stop")

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		select {
		case <-signals:
			// signal to stop test fixture
			cancel()
		case err = <-errChan:
			// stop waiting
		}
	}(cancel)

	// Wait for experiment completion
	//
	err = <-errChan
	if err != nil {
		// error during execution
		log.Fatalf("Experiment execution failed: %s", err)
	}

	// Print experiment results statistics
	//
	expt.PrintStatistics()

	fmt.Printf(">>> Start genome file:  %s\n", *genomePath)
	fmt.Printf(">>> Configuration file: %s\n", *contextPath)
	fmt.Printf(">>> Maze environment file: %s\n", *mazeConfigPath)

	// Save experiment data in native format
	//
	expResPath := fmt.Sprintf("%s/%s.dat", outDir, *experimentName)
	if expResFile, err := os.Create(expResPath); err != nil {
		log.Fatal("Failed to create file for experiment results", err)
	} else if err = expt.Write(expResFile); err != nil {
		log.Fatal("Failed to save experiment results", err)
	}

	// Save experiment data in Numpy NPZ format if requested
	//
	npzResPath := fmt.Sprintf("%s/%s.npz", outDir, *experimentName)
	if npzResFile, err := os.Create(npzResPath); err != nil {
		log.Fatalf("Failed to create file for experiment results: [%s], reason: %s", npzResPath, err)
	} else if err = expt.WriteNPZ(npzResFile); err != nil {
		log.Fatal("Failed to save experiment results as NPZ file", err)
	}
}
