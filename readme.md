# my-ls

`my-ls` is a custom implementation of the Unix `ls` command in Go. It provides functionality to list directory contents with various options for formatting and sorting.

## Features

- List directory contents in short or long format
- Recursive listing of subdirectories
- Show hidden files (files starting with a dot)
- Reverse sorting order
- Sort by modification time
- Colorized output for different file types
- Handling of device files in `/dev` directory

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

## Examples

1. List contents of the current directory:

./my-ls

2. Long format listing of a specific directory:

./my-ls -l /path/to/directory

3. Show all files, including hidden ones, sorted by modification time:

./my-ls -at

4. Recursive listing with reverse sorting:

./my-ls -Rr /path/to/directory


## Notes

- This implementation uses a fixed terminal width of 120 characters for formatting output. You can adjust this by changing the `terminalWidth` constant in the code.
- Color coding: 
- Blue: Directories
- Green: Executable files
- Cyan: Symbolic links
- Yellow: Device files
- Default color: Regular files
- The program handles device files specially when listing the `/dev` directory.

## Limitations

- The actual terminal width is not detected dynamically. If you need to change the assumed width, you'll need to modify the `terminalWidth` constant in the source code.
- Some advanced features of the standard `ls` command are not implemented.


