package maze

import (
	"testing"
	"math"
)

func TestPoint_Angle(t *testing.T) {
	var logError = func(angle, expected float64) {
		t.Errorf("Wrong angle found: %f, expected: %f\n", angle, expected)
	}
	// 0 degrees
	p := Point{1.0, 0.0}
	angle := p.Angle()
	if angle != 0 {
		logError(angle, 0.0)
	}
	// 90 degrees
	p.X = 0
	p.Y = 1.0
	angle = p.Angle()
	if angle != 90 {
		logError(angle, 90.0)
	}
	// 180 degrees
	p.X = -1.0
	p.Y = 0.0
	angle = p.Angle()
	if angle != 180 {
		logError(angle, 180.0)
	}
	// 270 degrees
	p.X = 0
	p.Y = -1.0
	angle = p.Angle()
	if angle != 270 {
		logError(angle, 270.0)
	}

	// 45 degrees
	p.X = 1.0
	p.Y = 1.0
	angle = p.Angle()
	if angle != 45 {
		logError(angle, 45.0)
	}
	// 135 degrees
	p.X = -1.0
	p.Y = 1.0
	angle = p.Angle()
	if angle != 135 {
		logError(angle, 135.0)
	}
	// 225 degrees
	p.X = -1.0
	p.Y = -1.0
	angle = p.Angle()
	if angle != 225 {
		logError(angle, 225.0)
	}
	// 315 degrees
	p.X = 1.0
	p.Y = -1.0
	angle = p.Angle()
	if angle != 315 {
		logError(angle, 315.0)
	}
}

func TestPoint_Rotate(t *testing.T) {
	p := Point{2.0, 1.0}

	p.Rotate(90.0, Point{1.0, 1.0})

	if p.X != 1.0 || p.Y != 2.0 {
		t.Error("Wrong position after rotation", p)
	}

	p.Rotate(180.0, Point{1.0, 1.0})

	if  1.0 - p.X > 0.00000001 || p.Y != 0.0 {
		t.Error("Wrong position after rotation", p)
	}
}

func TestPoint_Distance(t *testing.T) {
	p := Point{2.0, 1.0}
	p1 := Point{5.0, 1.0}

	d := p.Distance(p1)
	if d != 3 {
		t.Error("Wrong distance", d)
	}

	p2 := Point{5.0, 3.0}
	d = p.Distance(p2)
	if d != math.Sqrt(13.0){
		t.Error("Wrong distance", d)
	}
}

func TestLine_Intersection(t *testing.T) {
	l1 := Line{
		A:Point{1.0, 1.0},
		B:Point{5.0, 5.0},
	}
	l2 := Line{
		A:Point{1.0, 5.0},
		B:Point{5.0, 1.0},
	}

	// test intersection
	found, p := l1.Intersection(l2)
	if !found {
		t.Error("Lines intesect")
	}
	if p.X != 3.0 || p.Y != 3.0 {
		t.Error("Wrong intersection point found", p)
	}

	// test parallel
	l3 := Line{
		A:Point{2.0, 1.0},
		B:Point{6.0, 1.0},
	}
	found, p = l1.Intersection(l3)
	if found {
		t.Error("Parallel lines do not intesect")
	}
	if p.X != 0 || p.Y != 0 {
		t.Error("Wrong intersection point found", p)
	}

	// test no intersection by coordinates
	l4 := Line{
		A:Point{4.0, 4.0},
		B:Point{6.0, 1.0},
	}
	found, p = l1.Intersection(l4)
	if found {
		t.Error("The lines must not intesect")
	}
	if p.X != 0 || p.Y != 0 {
		t.Error("Wrong intersection point found", p)
	}
}

func TestLine_Distance(t *testing.T) {
	l := Line{
		A:Point{1.0, 1.0},
		B:Point{5.0, 1.0},
	}

	p := Point{4.0, 3.0}

	d := l.Distance(p)
	if d != 2.0 {
		t.Errorf("Wrong distance from line to point: %f, expected: %f\n", d, 2.0)
	}
}

func TestLine_Length(t *testing.T) {
	l := Line{
		A:Point{1.0, 1.0},
		B:Point{5.0, 1.0},
	}
	length := l.Length()
	if length != 4.0 {
		t.Errorf("Wrong line length: %f, expected: %f\n", length, 4.0)
	}
}