package maze

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

// The maximal allowed speed for maze agent
const maxAgentSpeed = 3.0

// Point the simple point class
type Point struct {
	X, Y float64
}

// ReadPoint reads Point from specified reader
func ReadPoint(lr io.Reader) (*Point, error) {
	point := &Point{}
	if _, err := fmt.Fscanf(lr, "%f %f", &point.X, &point.Y); err != nil {
		return nil, err
	}
	return point, nil
}

// Angle is to determine angle in degrees of vector defined by (0,0)->This Point.
// The angle is from 0 to 360 degrees anti-clockwise.
func (p Point) Angle() float64 {
	ang := math.Atan2(p.Y, p.X) / math.Pi * 180.0
	if ang < 0.0 {
		// lower quadrants (3 and 4)
		return ang + 360.0
	}
	return ang
}

// Rotate is to rotate this point around another point with given angle in degrees
func (p *Point) Rotate(angle float64, point Point) {
	rad := angle / 180.0 * math.Pi
	p.X -= point.X
	p.Y -= point.Y

	ox, oy := p.X, p.Y
	p.X = math.Cos(rad)*ox - math.Sin(rad)*oy
	p.Y = math.Sin(rad)*ox + math.Cos(rad)*oy

	p.X += point.X
	p.Y += point.Y
}

// Distance is to find distance between this point and another point
func (p Point) Distance(point Point) float64 {
	dx := point.X - p.X
	dy := point.Y - p.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// Line the simple line segment class, used for maze walls
type Line struct {
	A, B Point
}

// NewLine creates new line
func NewLine(a, b Point) Line {
	return Line{A: a, B: b}
}

// ReadLine reads line from specified reader
func ReadLine(lr io.Reader) (*Line, error) {
	a := Point{}
	b := Point{}
	if _, err := fmt.Fscanf(lr, "%f %f %f %f", &a.X, &a.Y, &b.X, &b.Y); err != nil {
		return nil, err
	}

	line := NewLine(a, b)
	return &line, nil
}

// Midpoint is to find midpoint of the line segment
func (l Line) Midpoint() Point {
	midpoint := Point{}
	midpoint.X = (l.A.X + l.B.X) / 2.0
	midpoint.Y = (l.A.Y + l.B.Y) / 2.0
	return midpoint
}

// Intersection calculates point of intersection between two line segments if it exists
func (l Line) Intersection(line Line) (bool, Point) {
	pt := Point{}
	A, B, C, D := l.A, l.B, line.A, line.B

	rTop := (A.Y-C.Y)*(D.X-C.X) - (A.X-C.X)*(D.Y-C.Y)
	rBot := (B.X-A.X)*(D.Y-C.Y) - (B.Y-A.Y)*(D.X-C.X)

	sTop := (A.Y-C.Y)*(B.X-A.X) - (A.X-C.X)*(B.Y-A.Y)
	sBot := (B.X-A.X)*(D.Y-C.Y) - (B.Y-A.Y)*(D.X-C.X)

	if rBot == 0 || sBot == 0 {
		// lines are parallel
		return false, pt
	}

	r := rTop / rBot
	s := sTop / sBot
	if r > 0 && r < 1 && s > 0 && s < 1 {
		pt.X = A.X + r*(B.X-A.X)
		pt.Y = A.Y + r*(B.Y-A.Y)

		return true, pt
	}
	return false, pt
}

// Distance is to find distance between line segment and the point
func (l Line) Distance(p Point) float64 {
	utop := (p.X-l.A.X)*(l.B.X-l.A.X) + (p.Y-l.A.Y)*(l.B.Y-l.A.Y)
	ubot := l.A.Distance(l.B)
	ubot *= ubot
	if ubot == 0.0 {
		return 0.0
	}

	u := utop / ubot
	if u < 0 || u > 1 {
		d1 := l.A.Distance(p)
		d2 := l.B.Distance(p)
		if d1 < d2 {
			return d1
		}
		return d2
	}
	point := Point{}
	point.X = l.A.X + u*(l.B.X-l.A.X)
	point.Y = l.A.Y + u*(l.B.Y-l.A.Y)
	return point.Distance(p)
}

// Length calculates the line segment length
func (l Line) Length() float64 {
	return l.A.Distance(l.B)
}

// Agent represents the maze navigator agent
type Agent struct {
	// The current location
	Location Point
	// The heading direction in degrees
	Heading float64
	// The speed of agent
	Speed float64
	// The angular velocity
	AngularVelocity float64
	// The radius of agent body
	Radius float64
	// The maximal range of range finder sensors
	RangeFinderRange float64

	// The angles of range finder sensors
	RangeFinderAngles []float64
	// The beginning angles for radar sensors
	RadarAngles1 []float64
	// The ending angles for radar sensors
	RadarAngles2 []float64

	// stores radar outputs
	Radar []float64
	// stores rangefinder outputs
	RangeFinders []float64
}

// NewAgent creates new Agent with default settings
func NewAgent() Agent {
	agent := Agent{
		Heading:          0.0,
		Speed:            0.0,
		AngularVelocity:  0.0,
		Radius:           8.0,
		RangeFinderRange: 100.0,
	}

	// define the range finder sensors
	agent.RangeFinderAngles = []float64{-90.0, -45.0, 0.0, 45.0, 90.0, -180.0}

	// define the radar sensors
	agent.RadarAngles1 = []float64{315.0, 45.0, 135.0, 225.0}
	agent.RadarAngles2 = []float64{405.0, 135.0, 225.0, 315.0}

	agent.RangeFinders = make([]float64, len(agent.RangeFinderAngles))
	agent.Radar = make([]float64, len(agent.RadarAngles1))

	return agent
}

// Environment the maze environment definition
type Environment struct {
	// The maze navigating agent
	Hero Agent
	// The maze line segments
	Lines []Line
	// The maze exit - goal
	MazeExit Point

	// The flag to indicate if exit was found
	ExitFound bool

	// The number of time steps to be executed during maze solving simulation
	TimeSteps int
	// The sample step size to determine when to collect subsequent samples during simulation
	SampleSize int

	// The range around maze exit point to test if agent coordinates is within to be considered as solved successfully (5.0 is good enough)
	ExitFoundRange float64

	// The initial distance of agent from exit
	initialDistance float64
}

// ReadEnvironment reads maze environment from the reader
func ReadEnvironment(ir io.Reader) (*Environment, error) {
	env := Environment{}
	env.Hero = NewAgent()
	env.Lines = make([]Line, 0)

	// Loop until file is finished, parsing each line
	scanner := bufio.NewScanner(ir)
	scanner.Split(bufio.ScanLines)
	index, numLines := 0, 0
	for scanner.Scan() {
		line := scanner.Text()
		lr := strings.NewReader(line)
		switch index {
		case 0: // read in how many line segments
			if _, err := fmt.Fscanf(lr, "%d", &numLines); err != nil {
				return nil, err
			}

		case 1: // read initial agent's location
			if loc, err := ReadPoint(lr); err != nil {
				return nil, err
			} else {
				env.Hero.Location = *loc
			}

		case 2: // read initial heading
			if _, err := fmt.Fscanf(lr, "%f", &env.Hero.Heading); err != nil {
				return nil, err
			}

		case 3: // read the maze exit location
			if loc, err := ReadPoint(lr); err != nil {
				return nil, err
			} else {
				env.MazeExit = *loc
			}

		default:
			// read maze line segments
			if line, err := ReadLine(lr); err != nil {
				return nil, err
			} else {
				env.Lines = append(env.Lines, *line)
			}
		}

		index++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if numLines != len(env.Lines) {
		return nil, errors.New(fmt.Sprintf("Expected: %d maze lines, but was read only: %d", numLines, len(env.Lines)))
	}

	// update sensors
	err := env.updateRangefinders()
	if err != nil {
		return nil, err
	}
	env.updateRadar()

	// find initial distance
	env.initialDistance = env.AgentDistanceToExit()

	return &env, err
}

// GetInputs create neural net inputs from maze agent sensors
func (e *Environment) GetInputs() ([]float64, error) {
	inputsSize := len(e.Hero.RangeFinders) + len(e.Hero.Radar) + 1
	inputs := make([]float64, inputsSize)
	// bias
	inputs[0] = 1.0

	// range finders
	i := 0
	for ; i < len(e.Hero.RangeFinders); i++ {
		inputs[1+i] = e.Hero.RangeFinders[i] / e.Hero.RangeFinderRange
		if math.IsNaN(inputs[1+i]) {
			return nil, errors.New("NAN in inputs from range finders")
		}
	}

	// radar
	for j := 0; j < len(e.Hero.Radar); j++ {
		inputs[i+j] = e.Hero.Radar[j]
		if math.IsNaN(inputs[i+j]) {
			return nil, errors.New("NAN in inputs from radar")
		}
	}

	return inputs, nil
}

// ApplyOutputs transform neural net outputs into angular velocity and speed
func (e *Environment) ApplyOutputs(o1, o2 float64) error {
	if math.IsNaN(o1) || math.IsNaN(o2) {
		return errors.New("OUTPUT is NAN")
	}

	e.Hero.AngularVelocity += o1 - 0.5
	e.Hero.Speed += o2 - 0.5

	// constraints of speed & angular velocity
	if e.Hero.Speed > maxAgentSpeed {
		e.Hero.Speed = maxAgentSpeed
	}
	if e.Hero.Speed < -maxAgentSpeed {
		e.Hero.Speed = -maxAgentSpeed
	}
	if e.Hero.AngularVelocity > maxAgentSpeed {
		e.Hero.AngularVelocity = maxAgentSpeed
	}
	if e.Hero.AngularVelocity < -maxAgentSpeed {
		e.Hero.AngularVelocity = -maxAgentSpeed
	}

	return nil
}

// Update does one time step of the simulation
func (e *Environment) Update() error {
	if e.ExitFound {
		return nil
	}

	// get horizontal and vertical velocity components
	vx := math.Cos(e.Hero.Heading/180.0*math.Pi) * e.Hero.Speed
	vy := math.Sin(e.Hero.Heading/180.0*math.Pi) * e.Hero.Speed

	if math.IsNaN(vx) {
		return errors.New("VX NAN")
	}
	if math.IsNaN(vy) {
		return errors.New("VY NAN")
	}

	// Update agent heading
	e.Hero.Heading += e.Hero.AngularVelocity
	if math.IsNaN(e.Hero.AngularVelocity) {
		return errors.New("HERO ANG VEL NAN")
	}

	if e.Hero.Heading > 360 {
		e.Hero.Heading -= 360
	}
	if e.Hero.Heading < 0 {
		e.Hero.Heading += 360
	}

	// Find next agent's location
	newLoc := Point{
		X: vx + e.Hero.Location.X,
		Y: vy + e.Hero.Location.Y,
	}
	if !e.testAgentCollision(newLoc) {
		e.Hero.Location.X = newLoc.X
		e.Hero.Location.Y = newLoc.Y
	}
	err := e.updateRangefinders()
	if err != nil {
		return err
	}
	e.updateRadar()

	// Test if update agent's position solved the maze
	e.ExitFound = e.testExitFoundByAgent()

	return nil
}

// testExitFoundByAgent is to test if agent location is within maze exit range
func (e *Environment) testExitFoundByAgent() bool {
	if e.ExitFound {
		return true
	}

	dist := e.AgentDistanceToExit()
	return dist < e.ExitFoundRange
}

// AgentDistanceToExit used for fitness calculations based on distance of maze Agent to the target maze exit
func (e *Environment) AgentDistanceToExit() float64 {
	return e.Hero.Location.Distance(e.MazeExit)
}

// update rangefinder sensors
func (e *Environment) updateRangefinders() error {
	// iterate through each sensor and find distance to maze lines with agent's range finder sensors
	for i := 0; i < len(e.Hero.RangeFinderAngles); i++ {
		// radians...
		rad := e.Hero.RangeFinderAngles[i] / 180.0 * math.Pi

		// project a point from the hero's location outwards
		projectionPoint := Point{
			X: e.Hero.Location.X + math.Cos(rad)*e.Hero.RangeFinderRange,
			Y: e.Hero.Location.Y + math.Sin(rad)*e.Hero.RangeFinderRange,
		}
		// rotate the projection point by the hero's heading
		projectionPoint.Rotate(e.Hero.Heading, e.Hero.Location)

		// create a line segment from the hero's location to projected
		projectionLine := Line{
			A: e.Hero.Location,
			B: projectionPoint,
		}

		// set range to max by default
		minRange := e.Hero.RangeFinderRange

		// now test against the environment to see if we hit anything
		for j := 0; j < len(e.Lines); j++ {
			found, intersection := e.Lines[j].Intersection(projectionLine)
			if found {
				// if so, then update the range to the distance
				foundRange := intersection.Distance(e.Hero.Location)

				// we want the closest intersection
				if foundRange < minRange {
					minRange = foundRange
				}
			}
		}

		if math.IsNaN(minRange) {
			return errors.New("RANGE is NAN")
		}
		e.Hero.RangeFinders[i] = minRange
	}
	return nil
}

// updateRadar is to update radar sensors
func (e *Environment) updateRadar() {
	target := e.MazeExit

	// rotate goal with respect to heading of agent to compensate agent's heading angle relative to zero heading angle
	target.Rotate(-e.Hero.Heading, e.Hero.Location)

	// translate with respect to location of agent to compensate agent's position relative to (0,0)
	target.X -= e.Hero.Location.X
	target.Y -= e.Hero.Location.Y

	// what angle is the vector between target & agent (agent is placed into (0,0) with zero heading angle due
	// to the affine transforms above)
	angle := target.Angle()

	// fire the appropriate radar sensor
	for i := 0; i < len(e.Hero.RadarAngles1); i++ {
		e.Hero.Radar[i] = 0.0

		if (angle >= e.Hero.RadarAngles1[i] && angle < e.Hero.RadarAngles2[i]) ||
			(angle+360.0 >= e.Hero.RadarAngles1[i] && angle+360.0 < e.Hero.RadarAngles2[i]) {
			e.Hero.Radar[i] = 1.0
		}
	}
}

// testAgentCollision is to see if provided new location hits anything in maze
func (e *Environment) testAgentCollision(loc Point) bool {
	for j := 0; j < len(e.Lines); j++ {
		if e.Lines[j].Distance(loc) < e.Hero.Radius {
			return true
		}
	}
	return false
}

// Stringer
func (e *Environment) String() string {
	str := fmt.Sprintf("MAZE\nHero at: %.1f, %.1f\n", e.Hero.Location.X, e.Hero.Location.Y)
	str += fmt.Sprintf("Exit at: %.1f, %.1f\n", e.MazeExit.X, e.MazeExit.Y)
	str += fmt.Sprintf("Initial distance from exit: %f, # of simulation steps: %d, path sampling size: %d \n",
		e.initialDistance, e.TimeSteps, e.SampleSize)
	str += "Lines:\n"
	for _, l := range e.Lines {
		str += fmt.Sprintf("\t[%.1f, %.1f] -> [%.1f, %.1f]\n", l.A.X, l.A.Y, l.B.X, l.B.Y)
	}
	return str
}
