package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"os"

	"github.com/namsral/flag"
)

func main() {
	args := os.Args
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "GLAB", flag.ExitOnError)
	file := fs.String("csv", "", "csv file with the results to parse")
	fs.Parse(args[1:])

	if *file == "" {
		log.Fatal("csv file name not provided.")
	}

	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("Error opening file: %s\n", *file)
	}
	r := csv.NewReader(bufio.NewReader(f))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(">>> %s <<<\n", record)
	}
}
