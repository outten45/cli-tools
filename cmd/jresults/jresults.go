package main

//go:generate go-bindata -o build_assets.go assets/

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/montanaflynn/stats"
	"github.com/namsral/flag"
)

func main() {
	// handle the arguments
	args := os.Args
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "JRES", flag.ExitOnError)
	file := fs.String("csv", "", "csv file with the results to parse")
	fs.Parse(args[1:])

	if *file == "" {
		log.Fatal("csv file name not provided.")
	}

	results := getResults(file)
	// fmt.Printf("%+v\n", results[0])

	d := getStats(results)
	// printResults(d)
	genHTML(d)
}

func genHTML(data *jstats) {
	file, err := Asset("assets/main.html")
	if err != nil {
		// Asset was not found.
		fmt.Printf("main.html wasn't not found: %+v\n", err)
	}

	t := template.New("Main HTML")
	t, _ = t.Parse(string(file))
	if err := t.Execute(os.Stdout, data); err != nil {
		log.Fatalf("Error with template: %+v\n", err)
	}
}

func printResults(data jstats) {
	for k, v := range data.Pages {
		mn, _ := stats.Mean(v.Elapsed)
		fmt.Printf("%s: \t %f (%d)\n", k, mn, len(v.Elapsed))
	}

	j, _ := json.Marshal(data.StatResults)
	fmt.Printf("json:\n%s\n", j)
}

func getResults(file *string) []*CSVResult {
	f, err := os.Open(*file)
	if err != nil {
		log.Fatalf("Error opening file: %s\n", *file)
	}

	results := []*CSVResult{}
	if err := gocsv.UnmarshalFile(f, &results); err != nil {
		log.Fatalf("Error with gocsv UnmarshalFile: %s\n", err)
	}
	return results
}

func setSuccessError(j *jstat, success string) {
	if success == "true" {
		j.Successes += 1
	} else {
		j.Errors += 1
	}
}

func getStats(results []*CSVResult) *jstats {
	j := &jstats{}
	groups := make(map[string]*jstat)
	for _, r := range results {
		if js, ok := groups[r.Label]; ok {
			js.Elapsed = append(js.Elapsed, stats.LoadRawData([]int{r.Elapsed})...)
			js.TimeStamps = append(js.TimeStamps, r.TimeStamp)
			groups[r.Label] = js
		} else {
			groups[r.Label] = &jstat{
				Label:      r.Label,
				TimeStamps: []int{r.TimeStamp},
				Elapsed:    stats.LoadRawData([]int{r.Elapsed}),
			}
		}
		js, _ := groups[r.Label]
		setSuccessError(js, r.Success)
	}
	j.Pages = groups
	j.GenStats()
	return j
}

type CSVResult struct {
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
// responseMessage  e.g. OK
// threadName - the name
// dataType - e.g. text
// success - true or false
// bytes - number of bytes in the sample
// latency - time to first response

type jstat struct {
	Label             string
	TimeStamps        []int
	Elapsed           stats.Float64Data
	Bytes             stats.Float64Data
	Successes         int
	Errors            int
	Samples           int
	Mean              float64
	StandardDeviation float64
	Median            float64
	Percent95         float64
	ErrorPercent      float64
}

type jstats struct {
	StartTime       time.Time
	Pages           map[string]*jstat
	StatResults     []*jstat
	StatResultsJSON string
}

// make a slice of jstat sort-able
type js []*jstat

func (a js) Len() int      { return len(a) }
func (a js) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a js) Less(i, j int) bool {
	return a[i].TimeStamps[0] < a[j].TimeStamps[0]
}

func (j *jstats) GenStats() []*jstat {
	m := make([]*jstat, 0)

	for _, js := range j.Pages {
		if mean, err := js.Elapsed.Mean(); err == nil {
			js.Mean = mean
		}
		if std, err := js.Elapsed.StandardDeviation(); err == nil {
			js.StandardDeviation = std
		}
		if med, err := js.Elapsed.Median(); err == nil {
			js.Median = med
		}
		if p95, err := js.Elapsed.Percentile(95.0); err == nil {
			js.Percent95 = p95
		}
		js.Samples = len(js.TimeStamps)
		js.ErrorPercent = (float64(js.Errors) / float64(js.Samples)) * 100.0
		m = append(m, js)
	}

	sort.Sort(js(m))
	j.StartTime = time.Unix(int64(m[0].TimeStamps[0]/1000), 0)
	j.StatResults = m

	jsonStr, _ := json.Marshal(m)
	j.StatResultsJSON = string(jsonStr)

	return m
}
