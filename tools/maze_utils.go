// The package holds variety of tools and utilities for data pre-/post-processing and results visualization.
package main

import (
	"io"
	"github.com/yaricom/goNEAT_NS/experiments/maze"
	"github.com/fogleman/gg"
	"image/color"
	"flag"
	"os"
	"log"
	"fmt"
	"math/rand"
	"image"
	"path"
)


// Draws maze agents records
func plotAgentsRecords(rs *maze.RecordStore, maze *maze.Environment, best_threshold float64, dc *gg.Context) {

	//plotAgentsRecordsByAge(records, dc)
	plotAgentsRecordsBySpecies(rs, maze, best_threshold, dc)

}

func plotAgentsRecordsByAge(records *maze.RecordStore, dc *gg.Context) {
	// find age range
	max_age, max_x, max_y := 0, 0.0, 0.0
	for _, r := range records.Records {
		if r.SpeciesAge > max_age {
			max_age = r.SpeciesAge
		}
		if r.X > max_x {
			max_x = r.X
		}
		if r.Y > max_y {
			max_y = r.Y
		}
	}
	fmt.Printf("The oldest age: %d\n", max_age)

	// build color scale
	c_scale := gg.NewLinearGradient(0, 0, float64(max_age), 0)

	c_scale.AddColorStop(0, color.RGBA{51, 153, 255, 255})
	c_scale.AddColorStop(1.0, color.RGBA{51, 0, 153, 255})


	// draw records
	for _, r := range records.Records {
		dc.DrawCircle(r.X, r.Y, 2.0)
		dc.SetColor(c_scale.ColorAt(r.SpeciesAge, 0))
		dc.Fill()
	}

	dc.SetFillStyle(c_scale)
	dc.DrawRectangle(5, float64(dc.Height() - 10), float64(dc.Width() - 10), 5)
	dc.Fill()
}

func plotAgentsRecordsBySpecies(records *maze.RecordStore, env *maze.Environment, best_threshold float64, dc *gg.Context) {
	max_id := 0
	for _, r := range records.Records {
		if r.SpeciesID > max_id {
			max_id = r.SpeciesID
		}
	}

	// find best species threshold
	dist_threshold := env.Hero.Location.Distance(env.MazeExit) * best_threshold


	// generate color palette and find best species (moved at least 2/3 fom start to exit)
	n_species, n_best_species := 0, 0
	colors := make([]color.Color, max_id + 1)
	sp_idx := make([]int, max_id + 1)
	best_sp_idx := make([]int, max_id + 1)
	for _, rec := range records.Records {
		if sp_idx[rec.SpeciesID] == 0 {
			sp_idx[rec.SpeciesID] = 1
			r, g, b := uint8(rand.Float64() * 255), uint8(rand.Float64() * 255), uint8(rand.Float64() * 255)
			colors[rec.SpeciesID] = color.RGBA{R:r, G:g, B:b, A:255}
			n_species++
		}
		if env.Hero.Location.Distance(maze.Point{rec.X, rec.Y}) >= dist_threshold {
			best_sp_idx[rec.SpeciesID]++
		}
	}


	// draw best species
	for i, v := range best_sp_idx {
		if v > 0 {
			n_best_species++
			plotSpecies(records, dc, i, colors)
		}
	}
	bounds := drawMaze(env, dc)
	drawMazeCaption(bounds, n_species, n_best_species, best_threshold, true, dc)

	fmt.Printf("total # of species: %d, # of the best species: %d\n", n_species, n_best_species)

	// draw worst species
	dc.Push()
	dc.Translate(0, 150)
	for i, v := range best_sp_idx {
		if v == 0 {
			plotSpecies(records, dc, i, colors)
		}
	}
	bounds = drawMaze(env, dc)
	drawMazeCaption(bounds, n_species, n_best_species, best_threshold, false, dc)
	dc.Pop()
}

func drawMazeCaption(bounds image.Rectangle, n_species, n_best_species int, b_threshold float64, best bool, dc *gg.Context) {
	x := float64(bounds.Max.X + 10)
	y := float64(bounds.Min.Y + bounds.Dy() / 2)

	var str string
	if best {
		str = fmt.Sprintf("fit >= %.1f", b_threshold)
	} else {
		str = fmt.Sprintf("fit < %.1f", b_threshold)
	}
	dc.SetColor(color.RGBA{0, 0, 102, 255})

	dc.DrawStringAnchored(str, x, y, 0, 0)
	if best {
		str = fmt.Sprintf("%d of %d", n_best_species, n_species)
	} else {
		str = fmt.Sprintf("%d of %d", n_species - n_best_species, n_species)
	}
	dc.DrawStringAnchored(str, x, y, 0, 1.0)
}

func drawMaze(maze *maze.Environment, dc *gg.Context) image.Rectangle {
	dc.SetColor(color.RGBA{0, 0, 102, 255})

	min_p, max_p := image.Pt(dc.Width(), dc.Height()), image.Pt(0, 0)

	min := func(p1, p2 image.Point) image.Point {
		if p1.X <= p2.X && p1.Y <= p2.Y {
			return p1
		} else {
			return p2
		}
	}
	max := func(p1, p2 image.Point) image.Point {
		if p1.X >= p2.X && p1.Y >= p2.Y {
			return p1
		} else {
			return p2
		}
	}

	// draw maze
	dc.Push()
	dc.SetLineWidth(3.0)
	dc.SetLineCap(gg.LineCapRound)
	for _, l := range maze.Lines {
		dc.DrawLine(l.A.X, l.A.Y, l.B.X, l.B.Y)
		dc.Stroke()

		min_p = min(min_p, image.Pt(int(l.A.X), int(l.A.Y)))
		min_p = min(min_p, image.Pt(int(l.B.X), int(l.B.Y)))

		max_p = max(max_p, image.Pt(int(l.A.X), int(l.A.Y)))
		max_p = max(max_p, image.Pt(int(l.B.X), int(l.B.Y)))
	}
	dc.Pop()

	// draw start point
	dc.Push()
	dc.SetLineWidth(2.0)
	dc.DrawCircle(maze.Hero.Location.X, maze.Hero.Location.Y, 4.0)
	dc.SetColor(color.RGBA{153, 255, 151, 255})
	dc.FillPreserve()
	dc.SetColor(color.White)
	dc.Stroke()

	// draw maze exit
	dc.DrawCircle(maze.MazeExit.X, maze.MazeExit.Y, 4.0)
	dc.SetColor(color.RGBA{255, 51, 0, 255})
	dc.FillPreserve()
	dc.SetColor(color.Gray{150})
	dc.Stroke()
	dc.Pop()

	return image.Rectangle{Min:min_p, Max:max_p}
}

func plotSpecies(records *maze.RecordStore, dc *gg.Context, speciesId int, colors []color.Color) {
	for _, r := range records.Records {
		if r.SpeciesID == speciesId {
			dc.DrawCircle(r.X, r.Y, 2.0)
			dc.SetColor(colors[r.SpeciesID])
			dc.Fill()
		}

	}
}

func drawMazeWithRecords(rec io.Reader, mr io.Reader, best_threshold float64, dc *gg.Context) error {
	env, err := maze.ReadEnvironment(mr)
	if err != nil {
		return err
	}
	rs := maze.RecordStore{}
	err = rs.Read(rec)
	if err != nil {
		return err
	}

	plotAgentsRecords(&rs, env, best_threshold, dc)

	return nil
}

func main() {
	var out_file_path = flag.String("out", "./out/out.png", "The PNG file to save visualization results.")
	var width = flag.Int("width", 400, "The canvas width for visualization")
	var height = flag.Int("height", 400, "The canvas height for visualization")
	var rec_path = flag.String("records", "", "The path to the file with agents recorded data")
	var maze_path = flag.String("maze", "", "The path to the maze environment config file")
	var best_threshold = flag.Float64("b_thresh", 0.8, "The minimal fitness of maze solving agent's species to be considered as the best ones.")

	flag.Parse()

	rand.Seed(int64(1042))

	dc := gg.NewContext(*width, *height)

	log.Printf("Loading records from: %s\n", *rec_path)

	if len(*rec_path) == 0 {
		log.Fatal("The records path not specified")
	}
	rec_file, err := os.Open(*rec_path)
	if err != nil {
		log.Fatalf("Failed to open agents records file: %s\n", *rec_path)
	}

	if len(*maze_path) == 0 {
		log.Fatal("The maze config file not set")
	}
	maze_file, err := os.Open(*maze_path)
	if err != nil {
		log.Fatalf("Failed to open maze config file: %s\n", *maze_path)
	}

	// set background
	dc.SetColor(color.White)
	dc.DrawRectangle(0, 0, float64(*width), float64(*height))
	dc.Fill()

	err = drawMazeWithRecords(rec_file, maze_file, *best_threshold, dc)
	if err != nil {
		log.Fatalf("Failed to render maze with agents, reason: %s\n", err)
	}

	// Check if output dir exists
	out_dir_path, _ := path.Split(*out_file_path)
	if _, err := os.Stat(out_dir_path); err != nil {
		// create output dir
		err = os.MkdirAll(out_dir_path, os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create output directory: ", err)
		}
	}

	dc.SavePNG(*out_file_path)
}
