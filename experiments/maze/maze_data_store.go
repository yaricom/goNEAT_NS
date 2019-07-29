package maze

import (
	"io"
	"encoding/gob"
	"errors"
	"fmt"
)

// The record holding info about individual maze agent performance at the end of simulation
type AgentRecord struct {
	// The ID of agent
	AgentID    int
	// The agent position at the end of simulation
	X, Y       float64
	// The agent fitness
	Fitness    float64
	// The flag to indicate whether agent reached maze exit
	GotExit    bool
	// The population generation when agent data was collected
	Generation int
	// The novelty value associated
	Novelty    float64

	// The ID of species to whom individual belongs
	SpeciesID  int
	// The age of species to whom individual belongs at time of recording
	SpeciesAge int
}

// The maze agent records storage
type RecordStore struct {
	// The array of agent records
	Records          []AgentRecord
	// The array of the solver agent path points
	SolverPathPoints []Point
}

// Writes record store to the provided writer
func (r *RecordStore) Write(w io.Writer) error {
	if len(r.Records) == 0 {
		return errors.New("No records to store")
	}
	// write the number of records and solver's path points
	fmt.Fprintf(w, "%d %d", len(r.Records), len(r.SolverPathPoints))
	// write records
	enc := gob.NewEncoder(w)
	for i := 0; i < len(r.Records); i++ {
		err := enc.Encode(r.Records[i])
		if err != nil {
			return err
		}
	}
	// write solver's path points
	for _, p := range r.SolverPathPoints {
		err := enc.Encode(p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Reads record store data from provided reader
func (rs *RecordStore) Read(r io.Reader) error {
	// read the number of records and solver path points
	var recNum, pathNum int
	fmt.Fscanf(r, "%d %d", &recNum, &pathNum)
	// read agents records
	rs.Records = make([]AgentRecord, recNum)
	dec := gob.NewDecoder(r)
	for i := 0; i < recNum; i++ {
		var ar AgentRecord
		err := dec.Decode(&ar)
		if err != nil {
			return err
		}
		rs.Records[i] = ar
	}
	// read solver path points
	if pathNum == 0 {
		return nil
	}
	rs.SolverPathPoints = make([]Point, pathNum)
	for i := 0; i < pathNum; i++ {
		var p Point
		err := dec.Decode(&p)
		if err != nil {
			return err
		}
		rs.SolverPathPoints[i] = p
	}
	return nil
}