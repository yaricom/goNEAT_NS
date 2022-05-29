package neatns

import (
	"encoding/json"
	"errors"
	"io"
)

// DumpNoveltyPoints dumps collected novelty points to the provided writer as JSON
func (a *NoveltyArchive) DumpNoveltyPoints(w io.Writer) error {
	if len(a.NovelItems) == 0 {
		return errors.New("no novel items to print")
	}
	return printNovelItems(a.NovelItems, w)
}

// DumpFittest dumps collected novelty points of individuals with maximal fitness found during evolution
func (a *NoveltyArchive) DumpFittest(w io.Writer) error {
	if len(a.FittestItems) == 0 {
		return errors.New("no fittest items to print")
	}
	return printNovelItems(a.FittestItems, w)
}

func printNovelItems(items []*NoveltyItem, w io.Writer) error {
	if data, err := json.Marshal(items); err != nil {
		return err
	} else if _, err = w.Write(data); err != nil {
		return err
	}
	return nil
}
