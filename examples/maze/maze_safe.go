package maze

import (
	"context"
	"errors"
	"fmt"
	"github.com/yaricom/goNEAT/v4/experiment"
	"github.com/yaricom/goNEAT/v4/experiment/utils"
	"github.com/yaricom/goNEAT/v4/neat"
	"github.com/yaricom/goNEAT/v4/neat/genetics"
	"github.com/yaricom/goNEAT/v4/neat/network"
	"github.com/yaricom/goNEAT_NS/v4/neatns"
	"math"
	"os"
)

// Implementation of the coevolution strategy SAFE (solution and fitness evolution) implementing commensalistic
// coevolution of the two populations: population of agents-solvers and population of candidates in objective function.

type objFunctionCandidate struct {
	coefficients []float64
}

type objFuncEvolutionManager struct {
	// the configuration options
	opts *neat.Options
	// The seed genome for objective function population
	startGenome *genetics.Genome

	// The novelty archive for objective functions
	archive *neatns.NoveltyArchive
	// The population of candidates into objective functions
	population *genetics.Population

	// The best objective function candidate by fitness
	bestFitnessObjFunction *objFunctionCandidate
}

type safeSearchEvaluator struct {
	// The routine to manage evolution of population of candidates into objective functions
	objFuncEvolution *objFuncEvolutionManager

	// The output path to store execution results
	outputPath string
	// The maze seed environment
	mazeEnv *Environment

	// The target number of species to be maintained
	numSpeciesTarget int
	// The species compatibility threshold adjustment frequency
	compatAdjustFreq int

	objFuncByOrgID map[int]*objFunctionCandidate
}

// NewSafeNSEvaluator allows creating maze solving agent using SAFE commensalistic coevolution method.
// It will use provided MazeEnv to run simulation of the maze environment. The objFuncGenome provided defines
// the start genome to be used for evolution of population of candidates into objective functions.
// The numSpeciesTarget specifies the target number of species to maintain in the population.
// If the number of species differ from the numSpeciesTarget it
// will be automatically adjusted with compatAdjustFreq frequency, i.e., at each epoch % compatAdjustFreq == 0
func NewSafeNSEvaluator(out string, mazeEnv *Environment, objFuncGenome *genetics.Genome, objFuncOpts *neat.Options, numSpeciesTarget, compatAdjustFreq int) (experiment.GenerationEvaluator, experiment.TrialRunObserver) {
	opts := neatns.DefaultNoveltyArchiveOptions()
	archiveObjFunc := neatns.NewNoveltyArchive(archiveThresh, NoveltyMetric, opts)
	objFuncEvolution := &objFuncEvolutionManager{
		startGenome: objFuncGenome,
		archive:     archiveObjFunc,
		opts:        objFuncOpts,
	}
	evaluator := &safeSearchEvaluator{
		outputPath:       out,
		mazeEnv:          mazeEnv,
		objFuncEvolution: objFuncEvolution,
		numSpeciesTarget: numSpeciesTarget,
		compatAdjustFreq: compatAdjustFreq,
	}
	return evaluator, evaluator
}

func (e *safeSearchEvaluator) spawnObjFuncPopulation() {
	e.objFuncEvolution.population = nil

	neat.InfoLog("\n>>>>> Spawning new population of objective function candidates ")
	pop, err := genetics.NewPopulation(e.objFuncEvolution.startGenome, e.objFuncEvolution.opts)
	if err != nil {
		neat.InfoLog("Failed to spawn new population of objective function candidates from start genome")
		return
	} else {
		neat.InfoLog("OK <<<<<")
	}
	neat.InfoLog(">>>>> Verifying spawned population of objective function candidates ")
	_, err = pop.Verify()
	if err != nil {
		neat.ErrorLog("\n!!!!! Population verification failed !!!!!")
		return
	} else {
		neat.InfoLog("OK <<<<<")
	}
	e.objFuncEvolution.population = pop
}

func (e *safeSearchEvaluator) TrialRunStarted(trial *experiment.Trial) {
	opts := neatns.DefaultNoveltyArchiveOptions()
	opts.KNNNoveltyScore = 10
	trialSim = mazeSimResults{
		trialID: trial.Id,
		records: new(RecordStore),
		archive: neatns.NewNoveltyArchive(archiveThresh, NoveltyMetric, opts),
	}
	// initialize map with objective function candidates
	e.objFuncByOrgID = make(map[int]*objFunctionCandidate)

	// spawn new population of objective function candidates
	e.spawnObjFuncPopulation()
}

func (e *safeSearchEvaluator) TrialRunFinished(_ *experiment.Trial) {
	// the last epoch executed
	e.storeRecorded()
}

func (e *safeSearchEvaluator) EpochEvaluated(_ *experiment.Trial, _ *experiment.Generation) {
	// just stub
}

func (e *safeSearchEvaluator) GenerationEvaluate(ctx context.Context, pop *genetics.Population, epoch *experiment.Generation) error {
	// check that population of candidates for objective function exists
	if e.objFuncEvolution.population == nil {
		return errors.New("no population of candidates for objective function found in every generation")
	}

	options, ok := neat.FromContext(ctx)
	if !ok {
		return neat.ErrNEATOptionsNotFound
	}
	// Evaluate each organism on a test
	for i, org := range pop.Organisms {
		res, err := e.solverEvaluate(org, pop, epoch)
		if err != nil {
			return err
		}
		// store fitness based on objective proximity for statistical purposes
		if org.Data == nil {
			neat.ErrorLog(fmt.Sprintf("Novelty point not found for organism: %s", org))
			pop.Organisms[i].Fitness = 0.0
		} else {
			pop.Organisms[i].Fitness = extractDistanceToExit(org.Data)
		}

		if res && (epoch.Champion == nil || org.Fitness > epoch.Champion.Fitness) {
			epoch.Solved = true
			epoch.WinnerNodes = len(org.Genotype.Nodes)
			epoch.WinnerGenes = org.Genotype.Extrons()
			epoch.WinnerEvals = trialSim.individualsCounter
			epoch.Champion = org
		}
	}

	// Fill statistics about current epoch
	epoch.FillPopulationStatistics(pop)

	// Only print to file every print_every generation
	if epoch.Solved || epoch.Id%options.PrintEvery == 0 || epoch.Id == options.NumGenerations-1 {
		if _, err := utils.WritePopulationPlain(e.outputPath, pop, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump population, reason: %s\n", err))
			return err
		}
	}

	if epoch.Solved {
		// print winner organism
		org := epoch.Champion
		utils.PrintActivationDepth(org, true)

		// print the objective function candidate coefficients
		if objFunc, ok := e.objFuncByOrgID[org.Genotype.Id]; ok {
			neat.InfoLog(fmt.Sprintf("\nThe solver's objective function: %s\n", objFunc))
		} else {
			neat.ErrorLog("The solver's objective function not found!!!!")
		}

		genomeFile := "maze_safe_winner"
		// Prints the winner organism's Genome to the file!
		if orgPath, err := utils.WriteGenomePlain(genomeFile, e.outputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's genome, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's genome dumped to: %s\n", epoch.Id, orgPath))
		}

		// Prints the winner organism's Phenotype to the Cytoscape JSON file!
		if orgPath, err := utils.WriteGenomeCytoscapeJSON(genomeFile, e.outputPath, org, epoch); err != nil {
			neat.ErrorLog(fmt.Sprintf("Failed to dump winner organism's phenome Cytoscape JSON graph, reason: %s\n", err))
		} else {
			neat.InfoLog(fmt.Sprintf("Generation #%d winner's phenome Cytoscape JSON graph dumped to: %s\n",
				epoch.Id, orgPath))
		}
	} else if epoch.Id < options.NumGenerations-1 {
		// adjust archive settings
		trialSim.archive.EndOfGeneration()
		// evaluate solvers fitness scores for the next epoch
		err := e.evaluateSolvers(pop, e.objFuncEvolution.population)
		if err != nil {
			neat.ErrorLog("Failed to evaluate solvers population fitness scores")
			return err
		}
		speciesCount := len(pop.Species)

		// adjust species count by keeping it constant
		adjustSpeciesNumber(speciesCount, epoch.Id, e.compatAdjustFreq, e.numSpeciesTarget, options)

		// evaluate fitness of candidates for objective functions for the next epoch
		e.objFuncEvolution.archive.EvaluatePopulationNovelty(e.objFuncEvolution.population, true)

		neat.InfoLog(fmt.Sprintf("%d species -> %d organisms [compatibility threshold: %.1f, target: %d]\n",
			speciesCount, len(pop.Organisms), options.CompatThreshold, e.numSpeciesTarget))
		neat.InfoLog(fmt.Sprintf("Best objective function candidate: %s", e.objFuncEvolution.bestFitnessObjFunction))
	}

	return nil
}

// evaluateSolvers is to evaluate population of solvers using provided population of the candidates for objective function
func (e *safeSearchEvaluator) evaluateSolvers(solversPop *genetics.Population, objFuncPop *genetics.Population) error {
	// evaluate candidates for objective functions
	functions := make([]*objFunctionCandidate, 0, len(objFuncPop.Organisms))
	for _, org := range objFuncPop.Organisms {
		if objF, err := e.objFuncEvaluate(org, objFuncPop); err != nil {
			neat.ErrorLog("Failed to evaluate objective function")
			return err
		} else {
			functions = append(functions, objF)
		}
	}

	// evaluate fitness of the solvers using candidates for objective functions
	maxFitness := 0.0
	for i, org := range solversPop.Organisms {
		if org.Data == nil {
			neat.ErrorLog("Skipping evaluation of solver organism without Novely Score")
			continue
		}
		fitness, objFunc := evaluateSolverFitness(org, functions)
		solversPop.Organisms[i].Fitness = fitness
		e.objFuncByOrgID[org.Genotype.Id] = objFunc
		if fitness > maxFitness {
			maxFitness = fitness
			e.objFuncEvolution.bestFitnessObjFunction = objFunc
		}
	}
	return nil
}

func (e *safeSearchEvaluator) objFuncEvaluate(org *genetics.Organism, pop *genetics.Population) (*objFunctionCandidate, error) {
	// get Organism phenotype's network depth
	phenotype, err := org.Phenotype()
	if err != nil {
		return nil, err
	}
	netDepth, err := phenotype.MaxActivationDepthWithCap(0) // The max depth of the network to be activated
	if err != nil {
		neat.DebugLog(fmt.Sprintf(
			"Failed to estimate maximal depth of the objective function candidate's network. Using default depth: %d", netDepth))
		neat.DebugLog(fmt.Sprintf("Genome: %s", org.Genotype))
	}
	neat.DebugLog(fmt.Sprintf("Objective function candidate's Network depth: %d for organism: %d\n", netDepth, org.Genotype.Id))
	if netDepth == 0 {
		neat.DebugLog(fmt.Sprintf("ALERT: Objective function candidate's Network depth is ZERO for Genome: %s", org.Genotype))
	}

	inputs := []float64{0.5}
	if err = phenotype.LoadSensors(inputs); err != nil {
		neat.ErrorLog("Failed to load sensors for objective function candidate")
		return nil, err
	}
	if _, err := phenotype.ForwardSteps(netDepth); err != nil && !errors.Is(err, network.ErrNetExceededMaxActivationAttempts) {
		neat.ErrorLog(fmt.Sprintf("Failed to activate network of objective function candidate: %s", err))
		return nil, err
	}
	nItem := neatns.NewNoveltyItem()
	nItem.IndividualID = org.Genotype.Id
	nItem.Data = append(nItem.Data, phenotype.Outputs[0].Activation)
	nItem.Data = append(nItem.Data, phenotype.Outputs[1].Activation)

	// evaluate Novelty score of the organism
	org.Data = &genetics.OrganismData{Value: nItem}
	e.objFuncEvolution.archive.EvaluateIndividualNovelty(org, pop, false)

	function := &objFunctionCandidate{
		coefficients: []float64{phenotype.Outputs[0].Activation, phenotype.Outputs[1].Activation},
	}
	return function, nil
}

// Evaluates individual maze solver organism against maze environment and returns true if organism was able to solve maze by navigating to exit
func (e *safeSearchEvaluator) solverEvaluate(org *genetics.Organism, pop *genetics.Population, epoch *experiment.Generation) (bool, error) {
	// create record to store simulation results for organism
	record := AgentRecord{Generation: epoch.Id, AgentID: trialSim.individualsCounter}
	record.SpeciesID = org.Species.Id
	record.SpeciesAge = org.Species.Age

	// evaluate individual organism and get novelty point
	nItem, solved, err := mazeSimulationEvaluate(e.mazeEnv, org, &record, nil)
	if err != nil {
		if errors.Is(err, ErrOutputIsNaN) {
			// corrupted genome, but OK to continue evolutionary process
			return false, nil
		}
		return false, err
	}
	nItem.IndividualID = org.Genotype.Id
	org.Data = &genetics.OrganismData{Value: nItem} // store novelty item within organism data
	org.IsWinner = solved                           // store if maze was solved
	org.Error = 1 - nItem.Fitness                   // error value consider how far  we are from exit normalized to (0;1] range

	// calculate novelty of new individual within archive of known novel items
	if !solved {
		trialSim.archive.EvaluateIndividualNovelty(org, pop, false)
		record.Novelty = org.Data.Value.(*neatns.NoveltyItem).Novelty // put it to the record
	} else {
		// solution found - set to maximal possible value
		record.Novelty = math.MaxFloat64

		// run simulation to store solver path
		pathPoints := make([]Point, e.mazeEnv.TimeSteps)
		if _, _, err = mazeSimulationEvaluate(e.mazeEnv, org, nil, pathPoints); err != nil {
			neat.ErrorLog(fmt.Sprintf("Solver's path simulation failed: %s\n", err))
			return false, err
		}
		trialSim.records.SolverPathPoints = pathPoints
	}

	// add record
	trialSim.records.Records = append(trialSim.records.Records, record)

	// increment tested unique individuals counter
	trialSim.individualsCounter++

	// update fittest organisms list
	if err = trialSim.archive.UpdateFittestWithOrganism(org); err != nil {
		return false, err
	}
	return solved, nil
}

func (e *safeSearchEvaluator) storeRecorded() {
	// store recorded agents' performance
	recPath := fmt.Sprintf("%s/record.dat", utils.CreateOutDirForTrial(e.outputPath, trialSim.trialID))
	recFile, err := os.Create(recPath)
	if err == nil {
		err = trialSim.records.Write(recFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to store agents' data records, reason: %s\n", err))
	}

	// print collected novelty points from archive
	npPath := fmt.Sprintf("%s/novelty_archive_points.json", utils.CreateOutDirForTrial(e.outputPath, trialSim.trialID))
	npFile, err := os.Create(npPath)
	if err == nil {
		err = trialSim.archive.DumpNoveltyPoints(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print novelty points from archive, reason: %s\n", err))
	}

	// print novelty points with maximal fitness
	npPath = fmt.Sprintf("%s/fittest_novelty_archive_points.json", utils.CreateOutDirForTrial(e.outputPath, trialSim.trialID))
	npFile, err = os.Create(npPath)
	if err == nil {
		err = trialSim.archive.DumpFittest(npFile)
	}
	if err != nil {
		neat.ErrorLog(fmt.Sprintf("Failed to print fittest novelty points from archive, reason: %s\n", err))
	}
}

// evaluateFitness allows applying SAFE method to calculate organism fitness
func (f *objFunctionCandidate) evaluateFitness(solverOrg *genetics.Organism) float64 {
	item := solverOrg.Data.Value.(*neatns.NoveltyItem)
	d := extractDistanceToExit(solverOrg.Data)

	return f.coefficients[0]*d + f.coefficients[1]*item.Novelty
}

func (f *objFunctionCandidate) String() string {
	return fmt.Sprintf("[%.5f, %.5f]", f.coefficients[0], f.coefficients[1])
}

// evaluateSolverFitness is to evaluate fitness of the solverOrg using provided candidates for objective functions and
// return the maximal found fitness value and the best objective function candidate
func evaluateSolverFitness(solverOrg *genetics.Organism, candidates []*objFunctionCandidate) (float64, *objFunctionCandidate) {
	maxFitness := 0.0
	var bestObjFunction *objFunctionCandidate
	for _, objFunc := range candidates {
		fitness := objFunc.evaluateFitness(solverOrg)
		if fitness > maxFitness {
			maxFitness = fitness
			bestObjFunction = objFunc
		}
	}
	return maxFitness, bestObjFunction
}

func extractDistanceToExit(data *genetics.OrganismData) float64 {
	return data.Value.(*neatns.NoveltyItem).Fitness
}
