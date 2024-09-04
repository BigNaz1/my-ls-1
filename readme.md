# my-ls

`my-ls` is a custom implementation of the Unix `ls` command in Go. It provides functionality to list directory contents with various options for formatting and sorting.

## Features

- List directory contents in short or long format
- Recursive listing of subdirectories
- Show hidden files (files starting with a dot)
- Reverse sorting order
- Sort by modification time
- Colorized output for different file types

## Usage
my-ls [OPTIONS] [DIRECTORY...]

If no directory is specified, `my-ls` will list the contents of the current directory.

### Options

- `-l`: Use long listing format
- `-R`: List subdirectories recursively
- `-a`: Do not ignore entries starting with a dot
- `-r`: Reverse order while sorting
- `-t`: Sort by modification time

## Building

To build the project, ensure you have Go installed on your system, then run:
go build -o my-ls

This will create an executable named `my-ls` in your current directory.

## Running

After building, you can run the program with:
./my-ls [OPTIONS] [DIRECTORY...]

## Notes

- This implementation uses a fixed terminal width of 80 characters for formatting output.
- Color coding: 
- Blue: Directories
- Green: Executable files
- Cyan: Symbolic links
- Default color: Regular files