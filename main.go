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
	"time"
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

	// Get directories or files to list
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	// List files for each argument
	for i, arg := range args {
		if i > 0 {
			fmt.Println()
		}

		err := handleSingleFile(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "my-ls: cannot access '%s': %v\n", arg, err)
		}
	}
}

func handleSingleFile(path string) error {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return listFiles(path, false)
	}

	// It's a file, just print its name
	if longFormat {
		printLongFormat([]fs.FileInfo{fileInfo}, filepath.Dir(path))
	} else {
		fmt.Println(getColoredName(fileInfo, path))
	}
	return nil
}

func listFiles(dir string, isRecursive bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []fs.FileInfo
	for _, entry := range entries {
		info, err := os.Lstat(filepath.Join(dir, entry.Name()))
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
		printLongFormat(files, dir)
	} else {
		printShortFormat(files, dir)
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

func printShortFormat(files []fs.FileInfo, dir string) {
	var names []string
	maxLen := 1

	// Add . and .. if showAll is true
	if showAll {
		names = append(names, getColoredName(&dirInfo{"."}, dir))
		names = append(names, getColoredName(&dirInfo{".."}, dir))
	}

	// Collect the file names and calculate the maximum name length
	for _, file := range files {
		name := getColoredName(file, filepath.Join(dir, file.Name()))
		names = append(names, name)
		plainName := stripANSI(name)
		if len(plainName) > maxLen {
			maxLen = len(plainName)
		}
	}

	// Calculate the column width (maxLen + 2 for padding between columns)
	columnWidth := maxLen + 2
	width := 80
	numCols := width / columnWidth

	if numCols <= 0 {
		numCols = 1
	}

	// Print the files row by row
	currentColumn := 0
	for _, name := range names {
		fmt.Print(name)
		currentColumn += len(stripANSI(name)) + 2 // Account for name and spacing
		if currentColumn >= width {
			fmt.Println() // Move to the next line
			currentColumn = 0
		} else {
			// Add padding if not the last column
			fmt.Print(strings.Repeat(" ", columnWidth-len(stripANSI(name))))
		}
	}

	// If the last row was not full, add a newline
	if currentColumn != 0 {
		fmt.Println()
	}
}

func stripANSI(str string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansi.ReplaceAllString(str, "")
}

func getColoredName(file fs.FileInfo, path string) string {
	name := file.Name()
	if name == "." || name == ".." {
		return colorBlue + name + colorReset
	}
	name = filepath.Base(name)

	if file.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err == nil {
			return colorCyan + name + colorReset + " -> " + linkTarget
		}
		return colorCyan + name + colorReset
	} else if file.IsDir() {
		return colorBlue + name + colorReset
	} else if file.Mode()&0111 != 0 {
		return colorGreen + name + colorReset
	}
	return name
}

func getGroupName(gid uint32) string {
	group, err := user.LookupGroupId(strconv.Itoa(int(gid)))
	if err == nil {
		return group.Name
	}
	return strconv.FormatUint(uint64(gid), 10)
}

func formatTime(t time.Time) string {
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)
	if t.Before(sixMonthsAgo) || t.After(now) {
		return t.Format("Jan _2  2006")
	}
	return t.Format("Jan _2 15:04")
}

func printLongFormat(files []fs.FileInfo, dir string) {
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

		nameLen := len(stripANSI(getColoredName(file, filepath.Join(dir, file.Name()))))
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

		coloredName := getColoredName(file, filepath.Join(dir, file.Name()))

		size := strconv.FormatInt(file.Size(), 10)
		if file.Mode()&os.ModeSymlink != 0 {
			size = "0"
		}

		fmt.Printf("%s %*d %-*s %-*s %*s %s %s\n",
			file.Mode(),
			maxLinkLen, stat.Nlink,
			maxUserLen, username,
			maxGroupLen, groupname,
			maxSizeLen, size,
			formatTime(file.ModTime()),
			coloredName)
	}
}

// Add this helper type and its methods
type dirInfo struct {
	name string
}

func (d *dirInfo) Name() string       { return d.name }
func (d *dirInfo) IsDir() bool        { return true }
func (d *dirInfo) Mode() os.FileMode  { return os.ModeDir }
func (d *dirInfo) ModTime() time.Time { return time.Time{} }
func (d *dirInfo) Size() int64        { return 0 }
func (d *dirInfo) Sys() interface{}   { return nil }
