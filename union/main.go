package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func run() error {
	files, err := filepath.Glob("_???.json")
	if err != nil {
		return err
	}
	fmt.Println("NOT EMPTY RESULTS:", len(files))
	total := len(files)
	res := map[string][]string{}
	for current, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		var chunk map[string][]string
		if err := json.Unmarshal(data, &chunk); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		for code, ids := range chunk {
			res[code] = ids
		}
		fmt.Printf("%.2f%%\r", (float64(current)+1)/float64(total)*100)
	}
	fmt.Println()
	out, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return os.WriteFile("data.json", out, 0o644)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
