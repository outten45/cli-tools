package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/namsral/flag"
)

func getResults(file *string) []*Result {
	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("Error opening file: %s\n", *file)
	}

	results := []*Result{}
	if err := gocsv.UnmarshalFile(f, &results); err != nil {
		log.Fatalf("Error with gocsv UnmarshalFile: %s\n", err)
	}
	return results
}

func main() {
	args := os.Args
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "GLAB", flag.ExitOnError)
	file := fs.String("csv", "", "csv file with the results to parse")
	fs.Parse(args[1:])

	if *file == "" {
		log.Fatal("csv file name not provided.")
	}

	results := getResults(file)
	fmt.Printf("%+v\n", results[0])
}

type Result struct {
	TimeStamp    int    `csv:"timeStamp"`
	Elapsed      int    `csv:"elapsed"`
	Label        string `csv:"label"`
	ResponseCode string `csv:"responseCode"`
	Success      string `csv:"success"`
	Bytes        int    `csv:"bytes"`
	Latency      int    `csv:"latency"`
}

// timeStamp - in milliseconds since 1/1/1970
// elapsed - in milliseconds
// label - sampler label
// responseCode - e.g. 200, 404
// responseMessage - e.g. OK
// threadName
// dataType - e.g. text
// success - true or false
// bytes - number of bytes in the sample
// latency - time to first response
