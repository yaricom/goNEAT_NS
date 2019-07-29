package main

import (
	"flag"
	"time"
	"os"
	"github.com/yaricom/goNEAT/neat"
	"github.com/yaricom/goNEAT/neat/genetics"
	"github.com/yaricom/goNEAT/experiments"
	"fmt"
	"math/rand"
	"log"
	"github.com/yaricom/goNEAT_NS/experiments/maze"
	"strings"
)


// The experiment runner boilerplate code
func main() {
	var out_dir_path = flag.String("out", "./out", "The output directory to store results.")
	var context_path = flag.String("context", "./data/maze.neat", "The execution context configuration file.")
	var genome_path = flag.String("genome", "./data/mazestartgenes", "The seed genome to start with.")
	var maze_config_path = flag.String("maze", "./data/medium_maze.txt", "The maze environment configuration file.")
	var experiment_name = flag.String("experiment", "MAZENS", "The name of experiment to run. [MAZENS, MAZEOBJ]")
	var time_steps = flag.Int("timesteps", 400, "The number of time steps for maze simulation per organism.")
	var time_steps_sample = flag.Int("timesteps_sample", 1000, "The sample size to store agent path when doing maze simulation.")
	var species_target = flag.Int("species_target", 20, "The target number of species to maintain.")
	var species_compat_adjust_freq = flag.Int("species_adjust_freq", 10, "The frequency of species compatibility theshold adjustments when trying to maintain their number.")
	var trials_count = flag.Int("trials", 0, "The number of trials for experiment. Overrides the one set in configuration.")
	var log_level = flag.Int("log_level", -1, "The logger level to be used. Overrides the one set in configuration.")
	var exit_range = flag.Float64("exit_range", 5.0, "The range around maze exit point to test if agent coordinates is within to be considered as solved successfully")

	flag.Parse()

	// Seed the random-number generator with current time so that
	// the numbers will be different every time we run.
	rand.Seed(time.Now().Unix())

	// Load context configuration
	configFile, err := os.Open(*context_path)
	if err != nil {
		log.Fatal("Failed to open context configuration file: ", err)
	}
	context := neat.LoadContext(configFile)

	// Load Genome
	log.Printf("Loading start genome for %s experiment\n", *experiment_name)
	genomeFile, err := os.Open(*genome_path)
	if err != nil {
		log.Fatal("Failed to open genome file: ", err)
	}
	encoding := genetics.PlainGenomeEncoding
	if strings.HasSuffix(*genome_path, ".yml") {
		encoding = genetics.YAMLGenomeEncoding
	}
	decoder, err := genetics.NewGenomeReader(genomeFile, encoding)
	if err != nil {
		log.Fatal("Failed to create genome decoder: ", err)
	}
	start_genome, err := decoder.Read()
	if err != nil {
		log.Fatal("Failed to read start genome: ", err)
	}
	fmt.Println(start_genome)

	// Load maze environment
	log.Printf("Reading maze environment: %s\n", *maze_config_path)
	mazeFile, err := os.Open(*maze_config_path)
	var environment *maze.Environment
	if err == nil {
		environment, err = maze.ReadEnvironment(mazeFile)
		if environment != nil {
			environment.TimeSteps = *time_steps
			environment.SampleSize = *time_steps_sample
			environment.ExitFoundRange = *exit_range
		}
		log.Println(environment)
	}
	if err != nil {
		log.Fatal("Failed to read maze environment configuration: ", err)
	}

	// Check if output dir exists
	out_dir := *out_dir_path
	if _, err := os.Stat(out_dir); err == nil {
		// backup it
		back_up_dir := fmt.Sprintf("%s-%s", out_dir, time.Now().Format("2006-01-02T15_04_05"))
		// clear it
		err = os.Rename(out_dir, back_up_dir)
		if err != nil {
			log.Fatal("Failed to do previous results backup: ", err)
		}
	}
	// create output dir
	err = os.MkdirAll(out_dir, os.ModePerm)
	if err != nil {
		log.Fatal("Failed to create output directory: ", err)
	}

	// Override context configuration parameters with ones set from command line
	if *trials_count > 0 {
		context.NumRuns = *trials_count
	}
	if *log_level >= 0 {
		neat.LogLevel = neat.LoggerLevel(*log_level)
	}

	experiment := experiments.Experiment{
		Id:0,
		Trials:make(experiments.Trials, context.NumRuns),
	}
	var generationEvaluator experiments.GenerationEvaluator
	if *experiment_name == "MAZENS" {
		generationEvaluator = maze.MazeNoveltySearchEvaluator{
			OutputPath:out_dir,
			Environment:environment,
			NumSpeciesTarget:*species_target,
			CompatAdjustFreq:*species_compat_adjust_freq,
		}
	} else if *experiment_name == "MAZEOBJ" {
		generationEvaluator = maze.MazeObjectiveEvaluator{
			OutputPath:out_dir,
			Environment:environment,
			NumSpeciesTarget:*species_target,
			CompatAdjustFreq:*species_compat_adjust_freq,
		}
	} else {
		log.Fatalf("Unsupported experiment name requested: %s\n", *experiment_name)
	}

	err = experiment.Execute(context, start_genome, generationEvaluator)
	if err != nil {
		log.Fatalf("Failed to perform %s experiment! Reason: %s\n", *experiment_name, err)
	}

	// Find winner statistics
	experiment.PrintStatistics()

	fmt.Printf(">>> Start genome file:  %s\n", *genome_path)
	fmt.Printf(">>> Configuration file: %s\n", *context_path)
	fmt.Printf(">>> Maze environment file: %s\n", *maze_config_path)

	// Save experiment data
	expResPath := fmt.Sprintf("%s/%s.dat", out_dir, *experiment_name)
	expResFile, err := os.Create(expResPath)
	if err == nil {
		err = experiment.Write(expResFile)
	}
	if err != nil {
		log.Fatal("Failed to save experiment results", err)
	}
}
