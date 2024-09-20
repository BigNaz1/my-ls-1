package opls

import (
	"io/fs"
	"os"
	"sort"
	"strings"
	"syscall"
)

func SortFiles(files []fs.FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		// Handle . and .. special cases
		if files[i].Name() == "." {
			return true // . always comes first
		}
		if files[j].Name() == "." {
			return false
		}
		if files[i].Name() == ".." {
			return files[j].Name() != "."
		}
		if files[j].Name() == ".." {
			return false
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
				return majorI < majorJ != ReverseSort
			}
			return minorI < minorJ != ReverseSort
		}

		// Sort by modification time if SortByTime flag is set
		if SortByTime {
			if files[i].ModTime().Equal(files[j].ModTime()) {
				// If modification times are equal, sort by name
				return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name()) != ReverseSort
			}
			return files[i].ModTime().After(files[j].ModTime()) != ReverseSort
		}

		// Sort by name
		return strings.ToLower(files[i].Name()) < strings.ToLower(files[j].Name()) != ReverseSort
	})
}
