package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	// Parse command line arguments
	var path string
	var printFiles bool

	flag.StringVar(&path, "path", ".", "//путь к директории, которую надо вывести")
	flag.BoolVar(&printFiles, "f", false, "//вывести не только директории, но и файлы")
	flag.Parse()

	out := os.Stdout

	err := DirTree(out, path, printFiles, "")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

}

func DirTree(out *os.File, path string, printFiles bool, prefix string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var dirs []os.DirEntry
	var files []os.DirEntry

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else if printFiles {
			files = append(files, entry)
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() < dirs[j].Name()
	})

	for i, dir := range dirs {
		isLast := i == len(dirs)-1 && len(files) == 0

		// подразделы
		fmt.Fprintf(out, "%s", prefix)
		if isLast {
			fmt.Fprint(out, "└───")
			prefix += "\t"
		} else {
			fmt.Fprint(out, "├───")
			prefix += "│\t"
		}

		fmt.Fprintln(out, dir.Name())

		nextPath := filepath.Join(path, dir.Name())
		err := DirTree(out, nextPath, printFiles, prefix)
		if err != nil {
			return err
		}

		prefix = prefix[:len(prefix)-1]
	}

	if printFiles {
		for i, file := range files {
			isLast := i == len(files)-1

			fmt.Fprintf(out, "%s", prefix)
			if isLast {
				fmt.Fprint(out, "└───")
			} else {
				fmt.Fprint(out, "├───")
			}

			fmt.Fprintln(out, file.Name())
		}
	}

	return nil
}
