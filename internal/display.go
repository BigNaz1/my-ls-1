package opls

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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

func PrintShortFormat(files []fs.FileInfo, dir string) {
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

func PrintLongFormat(files []fs.FileInfo, dir string) {
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
