package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// Command line flags
var (
	longFormat  bool
	recursive   bool
	showAll     bool
	reverseSort bool
	sortByTime  bool
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorBlue   = "\033[1;34m"
	colorGreen  = "\033[1;32m"
	colorCyan   = "\033[1;36m"
	colorYellow = "\033[1;33m"
)

func init() {
	// Define command line flags
	flag.BoolVar(&longFormat, "l", false, "Use long listing format")
	flag.BoolVar(&recursive, "R", false, "List subdirectories recursively")
	flag.BoolVar(&showAll, "a", false, "Do not ignore entries starting with .")
	flag.BoolVar(&reverseSort, "r", false, "Reverse order while sorting")
	flag.BoolVar(&sortByTime, "t", false, "Sort by modification time")
}

func main() {
	flag.Parse()

	// Get directories to list
	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	// List files for each directory
	for i, dir := range dirs {
		if i > 0 {
			fmt.Println()
		}
		err := listFiles(dir, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "my-ls: cannot access '%s': %v\n", dir, err)
		}
	}
}

func listFiles(dir string, isRecursive bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []fs.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !showAll && strings.HasPrefix(info.Name(), ".") {
			continue
		}
		files = append(files, info)
	}

	sortFiles(files)

	// Print directory name if needed
	if recursive || isRecursive {
		if dir == "." {
			fmt.Printf(".:\n")
		} else if strings.HasPrefix(dir, "./") {
			fmt.Printf("%s:\n", dir)
		} else {
			fmt.Printf("./%s:\n", dir)
		}
	}

	// Display file list
	if longFormat {
		printLongFormat(files)
	} else {
		printShortFormat(files)
	}

	// Handle recursive listing
	if recursive {
		for _, file := range files {
			if file.IsDir() && file.Name() != "." && file.Name() != ".." {
				subdir := filepath.Join(dir, file.Name())
				fmt.Println()
				listFiles(subdir, true)
			}
		}
	}

	return nil
}

func sortFiles(files []fs.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		// Handle . and .. special cases
		if files[i].Name() == "." {
			return !reverseSort
		}
		if files[j].Name() == "." {
			return reverseSort
		}
		if files[i].Name() == ".." {
			return !reverseSort && files[j].Name() != "."
		}
		if files[j].Name() == ".." {
			return reverseSort || files[i].Name() == "."
		}

		// Sort by modification time if -t flag is set
		if sortByTime {
			if files[i].ModTime().Equal(files[j].ModTime()) {
				// If modification times are equal, sort by name
				if reverseSort {
					return files[i].Name() > files[j].Name()
				}
				return files[i].Name() < files[j].Name()
			}
			if reverseSort {
				return files[i].ModTime().Before(files[j].ModTime())
			}
			return files[i].ModTime().After(files[j].ModTime())
		}

		// Sort by name
		if reverseSort {
			return files[i].Name() > files[j].Name()
		}
		return files[i].Name() < files[j].Name()
	})
}

func printShortFormat(files []fs.FileInfo) {
	var names []string
	maxLen := 0

	// First, add . and .. if showAll is true
	if showAll {
		names = append(names, ".", "..")
	}

	// Then add the rest of the files
	for _, file := range files {
		name := getColoredName(file)
		names = append(names, name)
		if len(stripANSI(name)) > maxLen {
			maxLen = len(stripANSI(name))
		}
	}

	// Use a fixed terminal width
	termWidth := 80
	columnWidth := maxLen + 2 // Add 2 for spacing
	numCols := termWidth / columnWidth
	if numCols == 0 {
		numCols = 1
	}

	// Print files in columns
	for i, name := range names {
		plainName := stripANSI(name)
		padding := strings.Repeat(" ", columnWidth-len(plainName))
		fmt.Print(name + padding)
		if (i+1)%numCols == 0 || i == len(names)-1 {
			fmt.Println()
		}
	}
}

func stripANSI(str string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansi.ReplaceAllString(str, "")
}

func getColoredName(file fs.FileInfo) string {
	name := file.Name()
	if name == "." || name == ".." {
		return name
	}
	name = filepath.Base(name)

	if file.IsDir() {
		return colorBlue + name + colorReset
	} else if file.Mode()&0111 != 0 {
		return colorGreen + name + colorReset
	} else if file.Mode()&os.ModeSymlink != 0 {
		return colorCyan + name + colorReset
	}
	return name
}

func getGroupName(gid uint32) string {
	return strconv.FormatUint(uint64(gid), 10)
}

func printLongFormat(files []fs.FileInfo) {
	maxLinkLen, maxUserLen, maxGroupLen, maxSizeLen, maxNameLen := 0, 0, 0, 0, 0

	// First pass: calculate maximum lengths for each column
	for _, file := range files {
		stat := file.Sys().(*syscall.Stat_t)

		linkLen := len(strconv.Itoa(int(stat.Nlink)))
		if linkLen > maxLinkLen {
			maxLinkLen = linkLen
		}

		username := strconv.Itoa(int(stat.Uid))
		if u, err := user.LookupId(username); err == nil {
			username = u.Username
		}
		if len(username) > maxUserLen {
			maxUserLen = len(username)
		}

		groupname := getGroupName(stat.Gid)
		if len(groupname) > maxGroupLen {
			maxGroupLen = len(groupname)
		}

		sizeLen := len(strconv.FormatInt(file.Size(), 10))
		if sizeLen > maxSizeLen {
			maxSizeLen = sizeLen
		}

		nameLen := len(stripANSI(getColoredName(file)))
		if nameLen > maxNameLen {
			maxNameLen = nameLen
		}
	}

	// Second pass: print files with proper alignment
	for _, file := range files {
		stat := file.Sys().(*syscall.Stat_t)

		username := strconv.Itoa(int(stat.Uid))
		if u, err := user.LookupId(username); err == nil {
			username = u.Username
		}

		groupname := getGroupName(stat.Gid)

		coloredName := getColoredName(file)
		plainName := stripANSI(coloredName)
		namePadding := strings.Repeat(" ", maxNameLen-len(plainName))

		fmt.Printf("%s %*d %-*s %-*s %*d %s %s%s\n",
			file.Mode(),
			maxLinkLen, stat.Nlink,
			maxUserLen, username,
			maxGroupLen, groupname,
			maxSizeLen, file.Size(),
			file.ModTime().Format("Jan _2 15:04"),
			coloredName,
			namePadding)
	}
}
