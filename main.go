package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/jawher/mow.cli"
)

type FileCount struct {
	Path  string
	Count int
}

var counts []FileCount
var dirCounts map[string]int = map[string]int{}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32768)
	count := 0
	lineSep := []byte{'\n'}
	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}
		count += bytes.Count(buf[:c], lineSep)
		if err == io.EOF {
			break
		}
	}
	return count, nil
}

func countLinesinFile(filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	count, err := lineCounter(file)
	if err != nil {
		return 0, err
	}
	file.Close()
	return count, nil
}

func getWalkFunc(dir string) filepath.WalkFunc {
	return func(filepath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if info.Mode().IsRegular() != true {
			return nil
		}

		count, err := countLinesinFile(filepath)
		if err != nil {
			return err
		}
		counts = append(counts, FileCount{filepath, count})

		dirCounts[dir] += count

		return nil
	}

}

func setupCli() error {
	app := cli.App("lu", "display line usage statistics")
	app.Spec = "[-s] [-c] DIRS..."
	grandTotal := app.BoolOpt("c", false, "Display a grand total.")
	selectedOnly := app.BoolOpt("s", false, "Display an entry for each specified file.")
	dirs := app.Strings(cli.StringsArg{
		Name: "DIRS",
		Desc: "Directories to analyze",
	})

	app.Action = func() {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		for _, dir := range *dirs {
			var dirAbsolute string
			if filepath.IsAbs(dir) == true {
				dirAbsolute = dir
			} else {
				dirAbsolute = path.Join(wd, dir)
			}
			dirExists, err := exists(dirAbsolute)
			if err != nil {
				panic(err)
			}
			if dirExists {
				countLines := getWalkFunc(dir)
				err = filepath.Walk(dirAbsolute, countLines)
				if err != nil {
					panic(err)
				}
			}
		}

		total := 0
		if *selectedOnly {
			for dir, count := range dirCounts {
				total += count
				fmt.Printf("%d\t%s\n", count, dir)
			}
		} else {
			for _, count := range counts {
				total += count.Count
				fmt.Printf("%d\t%s\n", count.Count, count.Path)
			}
		}
		if *grandTotal {
			fmt.Printf("%d\ttotal\n", total)
		}
	}

	return app.Run(os.Args)
}

func main() {
	err := setupCli()
	if err != nil {
		panic(err)
	}
}
