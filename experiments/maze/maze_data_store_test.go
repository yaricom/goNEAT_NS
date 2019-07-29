package maze

import (
	"testing"
	"bytes"
)

func TestRecordStore_Write_Read(t *testing.T) {
	rs := new(RecordStore)
	rs.Records = []AgentRecord{
		{0, 1, 2, 4, false, 1, 0, 1, 1},
		{1, 10, 20, 40, false, 1, 0, 1, 1},
		{2, 11, 21, 41, false, 1, 0, 1, 1},
		{3, 12, 22, 42, true, 1, 0, 1, 1},
	}
	rs.SolverPathPoints = [] Point {
		{0, 1},
		{2, 3},
		{4, 5},
	}

	// the store medium
	var store bytes.Buffer

	// store records
	err := rs.Write(&store)
	if err != nil {
		t.Error(err)
	}

	// read records to the new store
	nrs := new(RecordStore)

	err = nrs.Read(&store)
	if err != nil {
		t.Error(err)
	}

	// check results
	for i := 0; i < len(rs.Records); i++ {
		if rs.Records[i].AgentID != nrs.Records[i].AgentID {
			t.Error("rs.Records[i].AgentID != nrs.Records[i].AgentID")
		}
		if rs.Records[i].X != nrs.Records[i].X {
			t.Error("rs.Records[i].X != nrs.Records[i].X")
		}
		if rs.Records[i].Y != nrs.Records[i].Y {
			t.Error("rs.Records[i].Y != nrs.Records[i].Y")
		}
		if rs.Records[i].Fitness != nrs.Records[i].Fitness {
			t.Error("rs.Records[i].Fitness != nrs.Records[i].Fitness")
		}
		if rs.Records[i].GotExit != nrs.Records[i].GotExit {
			t.Error("rs.Records[i].GotExit != nrs.Records[i].GotExit")
		}
	}

	for i := 0; i < len(rs.SolverPathPoints); i++ {
		if rs.SolverPathPoints[i].X != nrs.SolverPathPoints[i].X {
			t.Error("rs.SolverPathPoints[i].X != nrs.SolverPathPoints[i].X",
				rs.SolverPathPoints[i].X, nrs.SolverPathPoints[i].X)
		}
		if rs.SolverPathPoints[i].Y != nrs.SolverPathPoints[i].Y {
			t.Error("rs.SolverPathPoints[i].Y != nrs.SolverPathPoints[i].Y",
				rs.SolverPathPoints[i].Y, nrs.SolverPathPoints[i].Y)
		}
	}
}
