package opls

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func HandleSingleFile(path string) error {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return ListFiles(path, false)
	}

	// It's a file, just print its name
	if LongFormat {
		PrintLongFormat([]fs.FileInfo{fileInfo}, filepath.Dir(path))
	} else {
		fmt.Println(getColoredName(fileInfo, path))
	}
	return nil
}

func ListFiles(dir string, isRecursive bool) error {
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
		if !ShowAll && strings.HasPrefix(info.Name(), ".") {
			continue
		}
		files = append(files, info)
	}

	SortFiles(files)

	// Print directory name if needed
	if Recursive || isRecursive {
		if dir == "." {
			fmt.Printf(".:\n")
		} else if strings.HasPrefix(dir, "./") {
			fmt.Printf("%s:\n", dir)
		} else {
			fmt.Printf("./%s:\n", dir)
		}
	}

	// Display file list
	if LongFormat {
		PrintLongFormat(files, dir)
	} else {
		PrintShortFormat(files, dir)
	}

	// Handle recursive listing
	if Recursive {
		for _, file := range files {
			if file.IsDir() && file.Name() != "." && file.Name() != ".." {
				subdir := filepath.Join(dir, file.Name())
				fmt.Println()
				ListFiles(subdir, true)
			}
		}
	}

	return nil
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
