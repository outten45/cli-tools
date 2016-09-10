package main

//go:generate go-bindata -o build_assets.go assets/

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"

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
	// fmt.Printf("%+v\n", results[0])

	d := getStats(results)
	d.Stats()
	// printResults(d)
	genHTML(d)
}

func genHTML(data jstats) {
	file, err := Asset("assets/main.html")
	if err != nil {
		// Asset was not found.
		fmt.Printf("main.html wasn't not found: %+v\n", err)
	}

	t := template.New("Main HTML")
	t, _ = t.Parse(string(file))
	t.Execute(os.Stdout, data)
}

func printResults(data jstats) {
	for k, v := range data.Pages {
		mn, _ := stats.Mean(v)
		fmt.Printf("%s: \t %f (%d)\n", k, mn, len(v))
	}

	t, _ := stats.Mean(data.Total)
	fmt.Printf("Total Mean: %f\n", t)

	s := data.Stats()
	for _, b := range s {
		fmt.Printf(" jstat: %+v\n----------------------------------------\n", b)
	}
	j, _ := json.Marshal(s)
	fmt.Printf("json:\n%s\n", j)
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

type jstat struct {
	Label             string
	Size              int
	Mean              float64
	StandardDeviation float64
	Median            float64
	Percent95         float64
}

type jstats struct {
	Pages       map[string]stats.Float64Data
	Total       stats.Float64Data
	StatResults []jstat
}

// make a slice of jstat sortable
type js []jstat

func (a js) Len() int      { return len(a) }
func (a js) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a js) Less(i, j int) bool {
	return a[i].Label < a[j].Label
}

func (j *jstats) Stats() []jstat {
	m := make([]jstat, 0)

	for label, page := range j.Pages {
		mean, _ := page.Mean()
		std, _ := page.StandardDeviation()
		med, _ := page.Median()
		p95, _ := page.Percentile(95.0)
		s := jstat{
			Label:             label,
			Size:              page.Len(),
			Mean:              mean,
			StandardDeviation: std,
			Median:            med,
			Percent95:         p95,
		}
		m = append(m, s)
	}
	sort.Sort(js(m))
	j.StatResults = m
	return m
}
