package maze

import (
	"io"
	"encoding/gob"
	"errors"
)

// The record holding info about individual maze agent performance at the end of simulation
type AgentRecord struct {
	// The ID of agent
	AgentID int
	// The agent position at the end of simulation
	X, Y    float64
	// The agent fitness
	Fitness float64
	// The flag to indicate whether agent reached maze exit
	GotExit bool
}

// The maze agent records storage
type RecordStore struct {
	// The array of agent records
	Records []AgentRecord
}

// Writes record store to the provided writer
func (r *RecordStore) Write(w io.Writer) error {
	if len(r.Records) == 0 {
		return errors.New("No records to store")
	}
	enc := gob.NewEncoder(w)
	for i := 0; i < len(r.Records); i++ {
		err := enc.Encode(r.Records[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Reads record store data from provided reader
func (rs *RecordStore) Read(r io.Reader) error {
	rs.Records = make([]AgentRecord, 0)
	dec := gob.NewDecoder(r)
	for true {
		var ar AgentRecord
		err := dec.Decode(&ar)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		rs.Records = append(rs.Records, ar)
	}
	return nil
}