package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/montanaflynn/stats"
	"github.com/namsral/flag"
)

func main() {
	// handle the arguments
	args := os.Args
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "GLAB", flag.ExitOnError)
	file := fs.String("csv", "", "csv file with the results to parse")
	fs.Parse(args[1:])

	if *file == "" {
		log.Fatal("csv file name not provided.")
	}

	results := getResults(file)

	fmt.Printf("%+v\n", results[0])

	d := getStats(results)
	// v, _ := stats.Mean(d.Total)
	// fmt.Printf("data: %+v\n------\n%+v\n", d, v)
	// fmt.Printf("data: %+v\n\n", d.Pages["Home Page"])
	printResults(d)
}

func printResults(data jstats) {
	for k, v := range data.Pages {
		mn, _ := stats.Mean(v)
		fmt.Printf("%s: \t %f\n", k, mn)
	}

	t, _ := stats.Mean(data.Total)
	fmt.Printf("Total Mean: %f\n", t)
}

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

func getStats(results []*Result) jstats {
	pagesRaw := make(map[string][]int)
	totalRaw := make([]int, len(results))

	for i, r := range results {
		if _, ok := pagesRaw[r.Label]; !ok {
			pagesRaw[r.Label] = make([]int, 0)
		}
		pagesRaw[r.Label] = append(pagesRaw[r.Label], r.Elapsed)
		totalRaw[i] = r.Elapsed
	}

	pages := make(map[string]stats.Float64Data)
	for k, v := range pagesRaw {
		pages[k] = stats.LoadRawData(v)
	}
	total := stats.LoadRawData(totalRaw)
	return jstats{Pages: pages, Total: total}
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

type jstats struct {
	Pages map[string]stats.Float64Data
	Total stats.Float64Data
}
