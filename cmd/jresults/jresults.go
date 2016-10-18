package main

//go:generate go-bindata -o build_assets.go assets/

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gocarina/gocsv"
	"github.com/montanaflynn/stats"
	"github.com/namsral/flag"
)

var buildstamp = "Not specified"
var githash = "Not specified"
var debug = false

func main() {
	// handle the arguments
	args := os.Args
	fs := flag.NewFlagSetWithEnvPrefix(args[0], "JRES", flag.ExitOnError)
	csv := fs.String("csv", "", "jtl file (csv) with the results to parse")
	dbFile := fs.String("db", "~/.jresults.db", "boltdb database to store results that have been processed")
	debug := fs.Bool("verbose", false, "turn on debugging output")
	version := fs.Bool("version", false, "provide build information")
	fs.Parse(args[1:])

	if *version {
		fmt.Printf("Build Stamp: %s\n", buildstamp)
		fmt.Printf("   Git Hash: %s\n", githash)
		os.Exit(0)
	}

	if *csv == "" {
		log.Fatal("csv file name not provided.")
	}

	results := getResults(csv)
	d := getStats(results)
	if *debug == true {
		printResults(d)
	}
	genHTML(d)
	saveToDatabase(dbFile, d)
}

// Save to the BoltStorageService at the given file location.
func saveToDatabase(dbFile *string, d *jstats) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	*dbFile = strings.Replace(*dbFile, "~", usr.HomeDir, 1)
	db, err := bolt.Open(fmt.Sprintf("%s", *dbFile), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	storage := &BoltStorageService{DB: db, BucketName: []byte("results")}
	ids, _ := storage.AllIds()
	_, err = storage.Results(ids[0])
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("display: %+v\n", jss.Key())
	storage.SaveResults(d)
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

func printResults(data *jstats) {
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
		j.Successes++
	} else {
		j.Errors++
	}
}

func getStats(results []*CSVResult) *jstats {
	j := &jstats{}
	all := &jstat{Label: "Totals"}

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

		all.TimeStamps = append(all.TimeStamps, r.TimeStamp)
		all.Elapsed = append(all.Elapsed, stats.LoadRawData([]int{r.Elapsed})...)
	}

	j.Totals = all
	j.Pages = groups
	j.GenStats()
	j.GenTotals()

	return j
}

// CSVResult contains the rows located in the result from the jmeter tests.
type CSVResult struct {
	TimeStamp    int    `csv:"timeStamp"`
	Elapsed      int    `csv:"elapsed"`
	Label        string `csv:"label"`
	ResponseCode string `csv:"responseCode"`
	Success      string `csv:"success"`
	Bytes        int    `csv:"bytes"`
	Latency      int    `csv:"Latency"`
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
// Latency - time to first response

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
	Totals          *jstat
	StatResultsJSON string
}

// make a slice of jstat sort-able
type js []*jstat

func (a js) Len() int      { return len(a) }
func (a js) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a js) Less(i, j int) bool {
	return a[i].TimeStamps[0] < a[j].TimeStamps[0]
}

// GenStats updates calculations like mean, standard deviation, median and others
// on pages in jstats.
func (j *jstats) GenStats() {
	var m = make([]*jstat, 0)

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

	return
}

func (j *jstats) Key() string {
	return j.StartTime.Format(time.RFC3339)
}

// GenTotals creates the stats (mean, standard deviation and more) for the
// totals in jstats.
func (j *jstats) GenTotals() {

	if mean, err := j.Totals.Elapsed.Mean(); err == nil {
		j.Totals.Mean = mean
	}
	if std, err := j.Totals.Elapsed.StandardDeviation(); err == nil {
		j.Totals.StandardDeviation = std
	}
	if med, err := j.Totals.Elapsed.Median(); err == nil {
		j.Totals.Median = med
	}
	if p95, err := j.Totals.Elapsed.Percentile(95.0); err == nil {
		j.Totals.Percent95 = p95
	}
	j.Totals.Samples = len(j.Totals.TimeStamps)
	j.Totals.ErrorPercent = (float64(j.Totals.Errors) / float64(j.Totals.Samples)) * 100.0

}
