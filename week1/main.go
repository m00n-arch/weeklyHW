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

	flag.StringVar(&path, "path", ".", "путь к директории, которую надо вывести")
	flag.BoolVar(&printFiles, "f", false, "вывести не только директории, но и файлы")
	flag.Parse()

	out := os.Stdout

	err := DirTree(out, path, printFiles, "")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func DirTree(out *os.File, path string, printFiles bool, prefix string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// Отладочный вывод для проверки размера файла в каталоге
	if len(entries) > 0 && !entries[0].IsDir() {
		//	filePath := filepath.Join(path, entries[0].Name())
		// fileSize := getFileSize(filePath)
		//	fmt.Println("Размер файла:", fileSize)
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

		fmt.Fprintf(out, "%s", prefix)
		if isLast {
			fmt.Fprint(out, "└───")
			prefix += "\t"
		} else {
			fmt.Fprint(out, "├───")
			prefix += "|\t"
		}

		fmt.Fprintln(out, dir.Name())

		nextPath := filepath.Join(path, dir.Name())
		err := DirTree(out, nextPath, printFiles, prefix)
		if err != nil {
			return err
		}

		if len(files) == 0 && i == len(dirs)-1 {
			prefix = prefix[:len(prefix)-1]
		} else {
			prefix = prefix[:len(prefix)-2]
		}
	}

	for i, file := range files {
		isLast := i == len(files)-1

		fmt.Fprintf(out, "%s", prefix)
		if isLast {
			fmt.Fprint(out, "└─── ")
		} else {
			fmt.Fprint(out, "├─── ")
		}

		filePath := filepath.Join(path, file.Name())
		fileSize := getFileSize(filePath)
		fmt.Println("FilePath:", filePath) // Отладочный вывод
		fmt.Println("FileSize:", fileSize) // Отладочный вывод
		fmt.Fprintf(out, "%s (%s)\n", file.Name(), formatSize(fileSize))
	}
	return nil
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func formatSize(size int64) string {
	const (
		_        = iota
		kilobyte = 1 << (10 * iota)
		megabyte
		gigabyte
	)

	switch {
	case size < kilobyte:
		return fmt.Sprintf("%d B", size)
	case size < megabyte:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(kilobyte))
	case size < gigabyte:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(megabyte))
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(gigabyte))
	}
}
