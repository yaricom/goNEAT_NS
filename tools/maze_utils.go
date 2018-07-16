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
)


// Draws maze agents records
func plotAgentsRecords(r io.Reader, dc *gg.Context) error {
	records := maze.RecordStore{}
	err := records.Read(r)
	if err != nil {
		return err
	}

	for _, r := range records.Records {
		dc.DrawCircle(r.X, r.Y, 3.0)
		dc.SetColor(color.White)
		dc.Fill()
	}

	return nil
}

func drawMazeWithRecords(rec io.Reader, dc *gg.Context) error {
	return plotAgentsRecords(rec, dc)
}

func main() {
	var out_dir_path = flag.String("out", "./out/out.png", "The PNG file to save visualization results.")
	var width = flag.Int("width", 400, "The canvas width for visualization")
	var height = flag.Int("height", 400, "The canvas height for visualization")
	var rec_path = flag.String("records", "", "The path to the file with agents recorded data")

	flag.Parse()

	dc := gg.NewContext(*width, *height)

	log.Printf("Loading records from: %s\n", *rec_path)

	if len(*rec_path) == 0 {
		log.Fatal("The records path not specified")
	}
	rec_file, err := os.Open(*rec_path)
	if err != nil {
		log.Fatalf("Failed to open agents records file: %s\n", *rec_path)
	}

	err = drawMazeWithRecords(rec_file, dc)
	if err != nil {
		log.Fatalf("Failed to render maze with agents, reason: %s\n", err)
	}

	// Check if output dir exists
	if _, err := os.Stat(*out_dir_path); err == nil {
		// create output dir
		err = os.MkdirAll(*out_dir_path, os.ModePerm)
		if err != nil {
			log.Fatal("Failed to create output directory: ", err)
		}
	}

	dc.SavePNG(*out_dir_path)
}
