// The maze solving experiments based on NEAT methodology with Novelty Search and Fitness based optimization
package maze

import (
	"math"
)

// The simple point class
type Point struct {
	X, Y float64
}

// To determine angle in degrees of vector defined by (0,0)->This Point. The angle is from 0 to 360 degrees anti clockwise.
func (p Point) Angle() float64 {
	ang := math.Atan2(p.Y, p.X) / math.Pi * 180.0
	if ang < 0.0 {
		// lower quadrants (3 and 4)
		return ang + 360.0
	}
	return ang
}

// To rotate this point around another point with given angle in degrees
func (p *Point) Rotate(angle float64, point Point) {
	rad := angle / 180.0 * math.Pi
	p.X -= point.X
	p.Y -= point.Y

	ox, oy := p.X, p.Y
	p.X = math.Cos(rad) * ox - math.Sin(rad) * oy
	p.Y = math.Sin(rad) * ox + math.Cos(rad) * oy

	p.X += point.X
	p.Y += point.Y
}

// To find distance between this point and another point
func (p Point) Distance(point Point) float64 {
	dx := point.X - p.X
	dy := point.Y - p.Y
	return math.Sqrt(dx * dx + dy * dy)
}

// The simple line segment class, used for maze walls
type Line struct {
	A, B Point
}

// To create new line
func NewLine(a, b Point) Line {
	return Line{A:a, B:b}
}

// To find midpoint of the line segment
func (l Line) Midpoint() Point {
	midpoint := Point{}
	midpoint.X = (l.A.X + l.B.X) / 2.0
	midpoint.Y = (l.A.Y + l.B.Y) / 2.0
	return midpoint
}

// Returns point of intersection between two line segments if it exists
func (l Line) Intersection(line Line) (bool, Point) {
	pt := Point{}
	A, B, C, D := l.A, l.B, line.A, line.B

	rTop := (A.Y - C.Y) * (D.X - C.X) - (A.X - C.X) * (D.Y - C.Y)
	rBot := (B.X - A.X) * (D.Y - C.Y) - (B.Y - A.Y) * (D.X - C.X)

	sTop := (A.Y - C.Y) * (B.X - A.X) - (A.X - C.X) * (B.Y - A.Y)
	sBot := (B.X - A.X) * (D.Y - C.Y) - (B.Y - A.Y) * (D.X - C.X)

	if rBot == 0 || sBot == 0 {
		// lines are parallel
		return false, pt
	}

	r := rTop / rBot
	s := sTop / sBot
	if r > 0 && r < 1 && s > 0 && s < 1 {
		pt.X = A.X + r * (B.X - A.X)
		pt.Y = A.Y + r * (B.Y - A.Y)

		return true, pt
	}
	return false, pt
}

// To find distance between line segment and the point
func (l Line) Distance(p Point) float64 {
	utop := (p.X - l.A.X) * (l.B.X - l.A.X) + (p.Y - l.A.Y) * (l.B.Y - l.A.Y)
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
	point.X = l.A.X + u * (l.B.X - l.A.X)
	point.Y = l.A.Y + u * (l.B.Y - l.A.Y)
	return point.Distance(p)
}

// The line segment length
func (l Line) Length() float64 {
	return l.A.Distance(l.B)
}

// The class for the maze navigator agent
type Agent struct {
	// The current location
	Location          Point
	// The heading direction in degrees
	Heading           float64
	// The speed of agent
	Speed             float64
	// The angular velocity
	AngularVelocity   float64
	// The radius of agent body
	Radius            float64
	// The maximal range of range finder sensors
	RangeFinderRange  float64

	// The angles of range finder sensors
	RangeFinderAngles []float64
	// The beginning angles for radar sensors
	RadarAngles1      []float64
	// The ending angles for radar sensors
	RadarAngles2      []float64

	// stores radar outputs
	Radar             []float64
	// stores rangefinder outputs
	RangeFinders      []float64
}

// Creates new Agent with default settings
func NewAgent() Agent {
	agent := Agent{
		Heading:0.0,
		Speed:0.0,
		AngularVelocity:0.0,
		Radius:8.0,
		RangeFinderRange:100.0,
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