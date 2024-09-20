package main

import (
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

// Assumed terminal width (you can adjust this value)
const terminalWidth = 120

func main() {
	// Custom flag parsing
	args := parseFlags()

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

func parseFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") && len(arg) > 1 && arg[1] != '-' {
			for _, ch := range arg[1:] {
				switch ch {
				case 'l':
					longFormat = true
				case 'R':
					recursive = true
				case 'a':
					showAll = true
				case 'r':
					reverseSort = true
				case 't':
					sortByTime = true
				default:
					fmt.Fprintf(os.Stderr, "my-ls: invalid option -- '%c'\n", ch)
					os.Exit(1)
				}
			}
		} else {
			args = append(args, arg)
		}
	}
	if len(args) == 0 {
		args = []string{"."}
	}
	return args
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

		// For device files, sort by major and minor numbers
		if files[i].Mode()&os.ModeDevice != 0 && files[j].Mode()&os.ModeDevice != 0 {
			devI := files[i].Sys().(*syscall.Stat_t).Rdev
			devJ := files[j].Sys().(*syscall.Stat_t).Rdev
			majorI := int64(devI >> 8)
			minorI := int64(devI & 0xff)
			majorJ := int64(devJ >> 8)
			minorJ := int64(devJ & 0xff)

			if majorI != majorJ {
				return majorI < majorJ
			}
			return minorI < minorJ
		}

		// Sort by modification time if -t flag is set
		if sortByTime {
			if files[i].ModTime().Equal(files[j].ModTime()) {
				// If modification times are equal, sort by name
				if reverseSort {
					return strings.ToLower(files[i].Name()) > strings.ToLower(files[j].Name())
				}
				return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
			}
			if reverseSort {
				return files[i].ModTime().Before(files[j].ModTime())
			}
			return files[i].ModTime().After(files[j].ModTime())
		}

		// Sort by name
		if reverseSort {
			return strings.ToLower(files[i].Name()) > strings.ToLower(files[j].Name())
		}
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name())
	})
}

func printShortFormat(files []fs.FileInfo, dir string) {
	var names []string
	maxLen := 0

	for _, file := range files {
		name := getColoredName(file, filepath.Join(dir, file.Name()))
		names = append(names, name)
		nameLen := len(stripANSI(name))
		if nameLen > maxLen {
			maxLen = nameLen
		}
	}

	numCols, colWidth := calculateColumnWidths(names, terminalWidth, maxLen)

	// Print in columns
	for i, name := range names {
		fmt.Print(name)
		if i%numCols == numCols-1 || i == len(names)-1 {
			fmt.Println()
		} else {
			padding := colWidth - len(stripANSI(name))
			fmt.Print(strings.Repeat(" ", padding))
		}
	}
}

func calculateColumnWidths(names []string, termWidth, maxLen int) (int, int) {
	if maxLen == 0 {
		return 1, termWidth
	}

	numCols := termWidth / (maxLen + 1)
	if numCols == 0 {
		numCols = 1
	}

	colWidth := termWidth / numCols

	// Adjust number of columns if there's too much extra space
	for numCols > 1 && numCols*maxLen < termWidth-numCols*2 {
		numCols++
		colWidth = termWidth / numCols
	}

	// Adjust number of columns if they don't fit
	for numCols > 1 && colWidth < maxLen+1 {
		numCols--
		colWidth = termWidth / numCols
	}

	return numCols, colWidth
}

func stripANSI(str string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansi.ReplaceAllString(str, "")
}

func getColoredName(file fs.FileInfo, path string) string {
	name := file.Name()
	mode := file.Mode()

	if mode&os.ModeDevice != 0 {
		if mode&os.ModeCharDevice != 0 {
			return colorYellow + name + colorReset
		}
		return colorYellow + colorBold + name + colorReset
	}

	if mode.IsDir() {
		return colorBlue + name + colorReset
	}
	if mode&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err == nil {
			return colorCyan + name + colorReset + " -> " + linkTarget
		}
		return colorCyan + name + colorReset
	}
	if mode&0o111 != 0 {
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
	maxLinkLen, maxUserLen, maxGroupLen, maxSizeLen := 0, 0, 0, 0

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

		size := file.Size()
		if file.Mode()&os.ModeDevice != 0 {
			major := int64(stat.Rdev >> 8)
			minor := int64(stat.Rdev & 0xff)
			size = major*256 + minor
		}
		sizeLen := len(strconv.FormatInt(size, 10))
		if sizeLen > maxSizeLen {
			maxSizeLen = sizeLen
		}
	}

	for _, file := range files {
		stat := file.Sys().(*syscall.Stat_t)

		username := strconv.Itoa(int(stat.Uid))
		if u, err := user.LookupId(username); err == nil {
			username = u.Username
		}

		groupname := getGroupName(stat.Gid)

		size := file.Size()
		if file.Mode()&os.ModeDevice != 0 {
			major := int64(stat.Rdev >> 8)
			minor := int64(stat.Rdev & 0xff)
			size = major*256 + minor
		}

		fmt.Printf("%s %*d %-*s %-*s %*d %s %s\n",
			file.Mode(),
			maxLinkLen, stat.Nlink,
			maxUserLen, username,
			maxGroupLen, groupname,
			maxSizeLen, size,
			formatTime(file.ModTime()),
			getColoredName(file, filepath.Join(dir, file.Name())))
	}
}
