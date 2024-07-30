// The package holds variety of tools and utilities for data pre-/post-processing and results visualization.
package main

import (
	"flag"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/yaricom/goNEAT_NS/v4/examples/maze"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
)

// plotAgentsRecords draws maze agents records
func plotAgentsRecords(rs *maze.RecordStore, maze *maze.Environment, bestThreshold float64, byAge bool, dc *gg.Context) {
	if byAge {
		plotAgentsRecordsByAge(rs, dc)
	} else {
		plotAgentsRecordsBySpecies(rs, maze, bestThreshold, dc)
	}
}

func plotAgentsRecordsByAge(records *maze.RecordStore, dc *gg.Context) {
	// find age range
	maxAge, maxX, maxY := 0, 0.0, 0.0
	for _, r := range records.Records {
		if r.SpeciesAge > maxAge {
			maxAge = r.SpeciesAge
		}
		if r.X > maxX {
			maxX = r.X
		}
		if r.Y > maxY {
			maxY = r.Y
		}
	}
	fmt.Printf("The oldest age: %d\n", maxAge)

	// build color scale
	colorScale := gg.NewLinearGradient(0, 0, float64(maxAge), 0)

	colorScale.AddColorStop(0, color.RGBA{R: 51, G: 153, B: 255, A: 255})
	colorScale.AddColorStop(1.0, color.RGBA{R: 51, B: 153, A: 255})

	// draw records
	for _, r := range records.Records {
		dc.DrawCircle(r.X, r.Y, 2.0)
		dc.SetColor(colorScale.ColorAt(r.SpeciesAge, 0))
		dc.Fill()
	}

	dc.SetFillStyle(colorScale)
	dc.DrawRectangle(5, float64(dc.Height()-10), float64(dc.Width()-10), 5)
	dc.Fill()
}

func plotAgentsRecordsBySpecies(records *maze.RecordStore, env *maze.Environment, bestThreshold float64, dc *gg.Context) {
	maxID := 0
	for _, r := range records.Records {
		if r.SpeciesID > maxID {
			maxID = r.SpeciesID
		}
	}

	// find the best species threshold
	distThreshold := env.AgentDistanceToExit() * (1.0 - bestThreshold)

	// generate color palette and find the best species (moved at least 2/3 fom start to exit)
	numSpecies, numBestSpecies := 0, 0
	colors := make([]color.Color, maxID+1)
	spIdx := make([]int, maxID+1)
	bestSpIdx := make([]int, maxID+1)
	for _, rec := range records.Records {
		if spIdx[rec.SpeciesID] == 0 {
			spIdx[rec.SpeciesID] = 1
			r, g, b := uint8(rand.Float64()*255), uint8(rand.Float64()*255), uint8(rand.Float64()*255)
			colors[rec.SpeciesID] = color.RGBA{R: r, G: g, B: b, A: 255}
			numSpecies++
		}
		if env.MazeExit.Distance(maze.Point{X: rec.X, Y: rec.Y}) <= distThreshold {
			bestSpIdx[rec.SpeciesID]++
		}
	}

	// draw best species
	for i, v := range bestSpIdx {
		if v > 0 {
			numBestSpecies++
			plotSpecies(records, dc, i, colors)
		}
	}
	bounds := drawMaze(env, dc)
	drawMazeCaption(bounds, numSpecies, numBestSpecies, bestThreshold, true, dc)

	fmt.Println(bounds)

	fmt.Printf("total # of species: %d, # of the best species: %d\n", numSpecies, numBestSpecies)

	// draw the worst species
	dc.Push()
	dc.Translate(0, float64(bounds.Max.Y+10))
	for i, v := range bestSpIdx {
		if v == 0 {
			plotSpecies(records, dc, i, colors)
		}
	}
	bounds = drawMaze(env, dc)
	drawMazeCaption(bounds, numSpecies, numBestSpecies, bestThreshold, false, dc)
	dc.Pop()
}

func drawMazeCaption(bounds image.Rectangle, numSpecies, numBestSpecies int, bestThreshold float64, best bool, dc *gg.Context) {
	x := float64(bounds.Max.X + 10)
	y := float64(bounds.Min.Y + bounds.Dy()/2)

	var str string
	if best {
		str = fmt.Sprintf("fit >= %.2f", bestThreshold)
	} else {
		str = fmt.Sprintf("fit < %.2f", bestThreshold)
	}
	dc.SetColor(color.RGBA{B: 102, A: 255})

	dc.DrawStringAnchored(str, x, y, 0, 0)
	if best {
		str = fmt.Sprintf("%d of %d", numBestSpecies, numSpecies)
	} else {
		str = fmt.Sprintf("%d of %d", numSpecies-numBestSpecies, numSpecies)
	}
	dc.DrawStringAnchored(str, x, y, 0, 1.0)
}

func drawMaze(maze *maze.Environment, dc *gg.Context) image.Rectangle {
	minX, minY, maxX, maxY := float64(dc.Width()), float64(dc.Height()), 0.0, 0.0

	// draw maze
	dc.Push()
	dc.SetColor(color.RGBA{B: 102, A: 255})
	dc.SetLineWidth(3.0)
	dc.SetLineCap(gg.LineCapRound)
	for _, l := range maze.Lines {
		dc.DrawLine(l.A.X, l.A.Y, l.B.X, l.B.Y)
		dc.Stroke()

		minX = math.Min(minX, l.A.X)
		minX = math.Min(minX, l.B.X)
		minY = math.Min(minY, l.A.Y)
		minY = math.Min(minY, l.B.Y)

		maxX = math.Max(maxX, l.A.X)
		maxX = math.Max(maxX, l.B.X)
		maxY = math.Max(maxY, l.A.Y)
		maxY = math.Max(maxY, l.B.Y)
	}
	dc.Pop()

	// draw start point
	dc.Push()
	dc.SetLineWidth(2.0)
	dc.DrawCircle(maze.Hero.Location.X, maze.Hero.Location.Y, 4.0)
	dc.SetColor(color.RGBA{R: 153, G: 255, B: 151, A: 255})
	dc.FillPreserve()
	dc.SetColor(color.White)
	dc.Stroke()

	// draw maze exit
	dc.DrawCircle(maze.MazeExit.X, maze.MazeExit.Y, 4.0)
	dc.SetColor(color.RGBA{R: 255, G: 51, A: 255})
	dc.FillPreserve()
	dc.SetColor(color.Gray{Y: 150})
	dc.Stroke()
	dc.Pop()

	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

func plotSpecies(records *maze.RecordStore, dc *gg.Context, speciesID int, colors []color.Color) {
	for _, r := range records.Records {
		if r.SpeciesID == speciesID {
			dc.DrawCircle(r.X, r.Y, 2.0)
			dc.SetColor(colors[r.SpeciesID])
			dc.Fill()
		}

	}
}

func plotSolverPath(records *maze.RecordStore, dc *gg.Context, color color.Color) {
	for _, p := range records.SolverPathPoints {
		dc.DrawCircle(p.X, p.Y, 2.0)
		dc.SetColor(color)
		dc.Fill()
	}
}

func drawMazeWithRecords(rec io.Reader, mr io.Reader, bestThreshold float64, byAge bool, dc *gg.Context) error {
	env, err := maze.ReadEnvironment(mr)
	if err != nil {
		return err
	}
	rs := maze.RecordStore{}
	err = rs.Read(rec)
	if err != nil {
		return err
	}

	plotAgentsRecords(&rs, env, bestThreshold, byAge, dc)

	return nil
}

func drawMazeWithPath(rec io.Reader, mr io.Reader, dc *gg.Context) error {
	env, err := maze.ReadEnvironment(mr)
	if err != nil {
		return err
	}
	rs := maze.RecordStore{}
	err = rs.Read(rec)
	if err != nil {
		return err
	}

	// draw the agents path
	plotSolverPath(&rs, dc, color.RGBA{B: 251, A: 255})

	// draw maze
	drawMaze(env, dc)

	fmt.Printf("Path rendered for %d points", len(rs.SolverPathPoints))

	return nil
}

func main() {
	var outFilePath = flag.String("out", "./out/out.png", "The PNG file to save visualization results.")
	var width = flag.Int("width", 400, "The canvas width for visualization")
	var height = flag.Int("height", 400, "The canvas height for visualization")
	var recPath = flag.String("records", "", "The path to the file with agents recorded data")
	var mazePath = flag.String("maze", "", "The path to the maze environment config file")
	var bestThreshold = flag.Float64("b_thresh", 0.8, "The minimal fitness of maze solving agent's species to be considered as the best ones.")
	var operation = flag.String("operation", "draw_agents", "The name of operation to apply to the records [draw_agents, draw_path].")
	var groupByAge = flag.Bool("group_by_age", false, "The flag to indicate whether agent records should be grouped by age of species")
	var scale = flag.Float64("scale", 1.0, "The scale factor for produced graphics")

	flag.Parse()

	rand.Seed(int64(1042))

	log.Printf("Loading records from: %s\n", *recPath)

	if len(*recPath) == 0 {
		log.Fatal("The records path not specified")
	}
	recFile, err := os.Open(*recPath)
	if err != nil {
		log.Fatalf("Failed to open agents records file: %s\n", *recPath)
	}

	if len(*mazePath) == 0 {
		log.Fatal("The maze config file not set")
	}
	mazeFile, err := os.Open(*mazePath)
	if err != nil {
		log.Fatalf("Failed to open maze config file: %s\n", *mazePath)
	}

	contextWidth := float64(*width) * *scale
	contextHeight := float64(*height) * *scale
	dc := gg.NewContext(int(contextWidth), int(contextHeight))
	// scale renderings if needed
	if *scale != 1.0 {
		dc.Scale(*scale, *scale)
	}

	// set background
	dc.SetColor(color.White)
	dc.DrawRectangle(0, 0, float64(*width), float64(*height))
	dc.Fill()

	switch *operation {
	case "draw_agents":
		err = drawMazeWithRecords(recFile, mazeFile, *bestThreshold, *groupByAge, dc)
	case "draw_path":
		err = drawMazeWithPath(recFile, mazeFile, dc)
	default:
		log.Fatalf("Usupported drawing operation requested: %s", *operation)

	}

	if err != nil {
		log.Fatalf("Failed to render agents records, reason: %s\n", err)
	}

	// Check if output dir exists
	outDirPath, _ := path.Split(*outFilePath)
	if _, err = os.Stat(outDirPath); err != nil {
		// create output dir
		err = os.MkdirAll(outDirPath, os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create output directory: ", err)
		}
	}

	if err = dc.SavePNG(*outFilePath); err != nil {
		log.Fatalf("Failed to save into PNG, reason: %s\n", err)
	}
}
