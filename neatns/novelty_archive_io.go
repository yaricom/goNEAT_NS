package neatns

import (
	"errors"
	"fmt"
	"io"
)

// PrintNoveltyPoints prints collected novelty points to provided writer
func (a *NoveltyArchive) PrintNoveltyPoints(w io.Writer) error {
	if len(a.NovelItems) == 0 {
		return errors.New("no novel items to print")
	}
	for _, p := range a.NovelItems {
		str := p.String()
		if _, err := fmt.Fprintln(w, str); err != nil {
			return err
		}
	}
	return nil
}

// PrintFittest prints collected individuals with maximal fitness
func (a *NoveltyArchive) PrintFittest(w io.Writer) error {
	if len(a.FittestItems) == 0 {
		return errors.New("no fittest items to print")
	}
	for _, f := range a.FittestItems {
		str := f.String()
		if _, err := fmt.Fprintln(w, str); err != nil {
			return err
		}
	}
	return nil
}
