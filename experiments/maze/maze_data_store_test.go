package maze

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRecordStore_Write_Read(t *testing.T) {
	rs := new(RecordStore)
	rs.Records = []AgentRecord{
		{0, 1, 2, 4, false, 1, 0, 1, 1},
		{1, 10, 20, 40, false, 1, 0, 1, 1},
		{2, 11, 21, 41, false, 1, 0, 1, 1},
		{3, 12, 22, 42, true, 1, 0, 1, 1},
	}
	rs.SolverPathPoints = []Point{
		{0, 1},
		{2, 3},
		{4, 5},
	}

	// the store medium
	var store bytes.Buffer

	// store records
	err := rs.Write(&store)
	require.NoError(t, err, "failed to save records")

	// read records to the new store
	var nrs RecordStore
	err = nrs.Read(&store)
	require.NoError(t, err, "failed to read records")

	// check that saved records match original
	assert.ElementsMatch(t, rs.Records, nrs.Records, "wrong records saved")
	assert.ElementsMatch(t, rs.SolverPathPoints, nrs.SolverPathPoints, "wrong solver path points saved")
}
