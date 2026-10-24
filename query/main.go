package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

const _dataPath = "data.json"

type entry struct {
	emote string
	count int
}

func run() error {
	if len(os.Args) == 1 {
		return errors.New("Usage: query [query]")
	}
	if len(os.Args) > 2 {
		return errors.New("Too many arguments")
	}
	query := os.Args[1]

	data, err := os.ReadFile(_dataPath)
	if err != nil {
		return err
	}
	db := map[string][]string{}
	if err := json.Unmarshal(data, &db); err != nil {
		return err
	}

	entries := []entry{}
	for emote, ids := range db {
		if strings.Contains(strings.ToLower(emote), query) {
			entries = append(entries, entry{emote, len(ids)})
		}
	}
	slices.SortFunc(entries, func(i, j entry) int {
		return cmp.Compare(j.count, i.count)
	})
	for _, e := range entries {
		fmt.Printf("%-20s %d\n", e.emote, e.count)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
