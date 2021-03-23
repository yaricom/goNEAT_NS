package maze

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"strings"
	"testing"
)

func TestPoint_Angle(t *testing.T) {

	// 0 degrees
	p := Point{1.0, 0.0}
	angle := p.Angle()
	assert.Equal(t, 0.0, angle, "wrong angle")

	// 90 degrees
	p.X = 0
	p.Y = 1.0
	angle = p.Angle()
	assert.Equal(t, 90.0, angle, "wrong angle")

	// 180 degrees
	p.X = -1.0
	p.Y = 0.0
	angle = p.Angle()
	assert.Equal(t, 180.0, angle, "wrong angle")

	// 270 degrees
	p.X = 0
	p.Y = -1.0
	angle = p.Angle()
	assert.Equal(t, 270.0, angle, "wrong angle")

	// 45 degrees
	p.X = 1.0
	p.Y = 1.0
	angle = p.Angle()
	assert.Equal(t, 45.0, angle, "wrong angle")

	// 135 degrees
	p.X = -1.0
	p.Y = 1.0
	angle = p.Angle()
	assert.Equal(t, 135.0, angle, "wrong angle")

	// 225 degrees
	p.X = -1.0
	p.Y = -1.0
	angle = p.Angle()
	assert.Equal(t, 225.0, angle, "wrong angle")

	// 315 degrees
	p.X = 1.0
	p.Y = -1.0
	angle = p.Angle()
	assert.Equal(t, 315.0, angle, "wrong angle")
}

func TestPoint_Rotate(t *testing.T) {
	p := Point{2.0, 1.0}

	p.Rotate(90.0, Point{1.0, 1.0})

	assert.Equal(t, 1.0, p.X, "Wrong X coordinate after rotation")
	assert.Equal(t, 2.0, p.Y, "Wrong Y coordinate after rotation")

	p.Rotate(180.0, Point{1.0, 1.0})

	assert.InDelta(t, 1.0, p.X, 0.00000001, "Wrong X coordinate after rotation")
	assert.Equal(t, 0.0, p.Y, "Wrong Y coordinate after rotation")
}

func TestPoint_Distance(t *testing.T) {
	p := Point{2.0, 1.0}
	p1 := Point{5.0, 1.0}

	d := p.Distance(p1)
	assert.Equal(t, 3.0, d, "Wrong distance")

	p2 := Point{5.0, 3.0}
	d = p.Distance(p2)
	expected := math.Sqrt(13.0)
	assert.InDelta(t, expected, d, 0.00000001, "Wrong distance")
}

func TestReadPoint(t *testing.T) {
	str := "10 20"
	lr := strings.NewReader(str)

	point := ReadPoint(lr)
	assert.Equal(t, 10.0, point.X, "Point has wrong X")
	assert.Equal(t, 20.0, point.Y, "Point has wrong Y")
}

func TestLine_Intersection(t *testing.T) {
	l1 := Line{
		A: Point{1.0, 1.0},
		B: Point{5.0, 5.0},
	}
	l2 := Line{
		A: Point{1.0, 5.0},
		B: Point{5.0, 1.0},
	}

	// test intersection
	found, p := l1.Intersection(l2)
	assert.True(t, found, "Lines intersecting")
	assert.Equal(t, 3.0, p.X, "Wrong intersection point's X coordinate")
	assert.Equal(t, 3.0, p.Y, "Wrong intersection point's Y coordinate")

	// test parallel
	l3 := Line{
		A: Point{2.0, 1.0},
		B: Point{6.0, 1.0},
	}
	found, p = l1.Intersection(l3)
	assert.False(t, found, "Parallel lines do not intersect")
	assert.Equal(t, 0.0, p.X, "Wrong intersection point's X coordinate")
	assert.Equal(t, 0.0, p.Y, "Wrong intersection point's Y coordinate")

	// test no intersection by coordinates
	l4 := Line{
		A: Point{4.0, 4.0},
		B: Point{6.0, 1.0},
	}
	found, p = l1.Intersection(l4)
	assert.False(t, found, "The lines must not intersect")
	assert.Equal(t, 0.0, p.X, "Wrong intersection point's X coordinate")
	assert.Equal(t, 0.0, p.Y, "Wrong intersection point's Y coordinate")
}

func TestLine_Distance(t *testing.T) {
	l := Line{
		A: Point{1.0, 1.0},
		B: Point{5.0, 1.0},
	}

	p := Point{4.0, 3.0}

	d := l.Distance(p)
	assert.Equal(t, 2.0, d, "Wrong distance from line to point")
}

func TestLine_Length(t *testing.T) {
	l := Line{
		A: Point{1.0, 1.0},
		B: Point{5.0, 1.0},
	}
	length := l.Length()
	assert.Equal(t, 4.0, length, "Wrong line length")
}

func TestReadLine(t *testing.T) {
	str := "10 20 30 40"
	lr := strings.NewReader(str)

	line, err := ReadLine(lr)
	require.NoError(t, err, "failed to read line")
	assert.Equal(t, 10.0, line.A.X)
	assert.Equal(t, 20.0, line.A.Y)
	assert.Equal(t, 30.0, line.B.X)
	assert.Equal(t, 40.0, line.B.Y)
}

func TestReadEnvironment(t *testing.T) {
	mazeConfigPath := "../../data/medium_maze.txt"

	// open maze config file
	mazeFile, err := os.Open(mazeConfigPath)
	require.NoError(t, err, "failed to read maze file")

	env, err := ReadEnvironment(mazeFile)
	require.NoError(t, err, "failed to read environment")

	assert.Equal(t, 30.0, env.Hero.Location.X)
	assert.Equal(t, 22.0, env.Hero.Location.Y)
	assert.Len(t, env.Lines, 11, "wrong number of walls")
	assert.Equal(t, 270.0, env.MazeExit.X)
	assert.Equal(t, 100.0, env.MazeExit.Y)

	lines := []Line{
		{Point{5, 5}, Point{295, 5}},
		{Point{295, 5}, Point{295, 135}},
		{Point{295, 135}, Point{5, 135}},
		{Point{5, 135}, Point{5, 5}},
		{Point{241, 135}, Point{58, 65}},
		{Point{114, 5}, Point{73, 42}},
		{Point{130, 91}, Point{107, 46}},
		{Point{196, 5}, Point{139, 51}},
		{Point{219, 125}, Point{182, 63}},
		{Point{267, 5}, Point{214, 63}},
		{Point{271, 135}, Point{237, 88}},
	}
	assert.ElementsMatch(t, lines, env.Lines)
}
