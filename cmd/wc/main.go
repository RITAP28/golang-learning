// rebuilding the unix word-count tool
// run the following commands for testing this tool out:
// 1. cd cmd/wc/
// 2. go build main.go
// 3. ./main example.txt exampleTwo.txt exampleThree.txt

// sub-tasks:
// 1. read from a single file and print static text	-done
// 2. building the core couting logic				-done
// 3. supporting the command-line flags				-done
// 4. read from standard input (stdin)				-done
// 5. handling multiple files and the "total" row	-done
// 6. error handling and non-zero exit code			-done

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	// way of declaring variables in golang at one go
	var (
		bytesCount	int
		linesCount	int
		runesCount	int
		wordsCount	int
		hadError	bool
	)

	// declaring the flags at the top
	bytesPtr := flag.Bool("c", false, "count bytes")
	runesPtr := flag.Bool("m", false, "count runes")
	wordsPtr := flag.Bool("w", false, "count words")
	linesPtr := flag.Bool("l", false, "count lines")

	// parsing the flags, which populates flag.Args()
	flag.Parse()

	// fmt.Println("program name: ", os.Args[0])

	// gracefully return an error if no filename is provided with the command
	if len(flag.Args()) == 0 {
		// reading directly from the user input
		// in case no file is provided by the user to read from
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()

			linesCount = strings.Count(string(line), "\n")
			bytesCount += len(scanner.Bytes()) + 1
			runesCount += len([]rune(string(line))) + 1
			wordsCount += len(strings.Fields(string(line)))
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading stdin: ", err)
			return
		}
	}

	files := flag.Args()[0:]
	for index, file := range files {
		fmt.Printf("Index: %d, FileName: %s\n", index, file)

		// reading the file with os.ReadFile
		// could also have used os.Open(filename) and then read
		contentBytes, err := os.ReadFile(file)

		if err == io.EOF {
			hadError = true
			fmt.Fprintln(os.Stderr, "Error: ", err)
			continue
		}
		if err != nil {
			hadError = true
			fmt.Fprintln(os.Stderr, "Error: ", err)
			continue
		}

		// running the necessary calculations
		bytesCount = len(contentBytes)
		runesCount = len([]rune(string(contentBytes)))
		linesCount = strings.Count(string(contentBytes), "\n")
		wordsCount = len(strings.Fields(string(contentBytes)))

		// fmt.Printf("%d %d %d %s\n", linesCount, wordsCount, bytesCount, targetFileName)

		// by default, if nothing is mentioned, bytes, words and lines shall appear
		if !*bytesPtr && !*runesPtr && !*wordsPtr && !*linesPtr {
			*bytesPtr = true
			*wordsPtr = true
			*linesPtr = true
		}
	
		var output []string
		if *bytesPtr {
			output = append(output, fmt.Sprintf("%d", bytesCount))
		}
		if *runesPtr {
			output = append(output, fmt.Sprintf("%d", runesCount))
		}
		if *wordsPtr {
			output = append(output, fmt.Sprintf("%d", wordsCount))
		}
		if *linesPtr {
			output = append(output, fmt.Sprintf("%d", linesCount))
		}

		output = append(output, file)
		fmt.Println(strings.Join(output, " "))
	}

	// exiting with code 1
	if hadError {
		os.Exit(1)
	}
}
