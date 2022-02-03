package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/covince/covince-backend-v2/api"
	"github.com/covince/covince-backend-v2/covince"
	"github.com/covince/covince-backend-v2/perf"
)

func indexMutations(muts []string) map[string][]string {
	i := make(map[string][]string)
	for _, m := range muts {
		split := strings.Split(m, ":")
		gene := split[0]
		description := split[1]
		existing, ok := i[gene]
		if !ok {
			existing = make([]string, 0)
		}
		i[gene] = append(existing, description)
	}
	return i
}

func createRecordFromCsv(row []string) covince.Record {
	count, _ := strconv.Atoi(row[5])
	return covince.Record{
		Date:       row[0],
		Lineage:    row[1],
		PangoClade: row[2],
		Area:       row[3],
		Mutations:  indexMutations(strings.Split(row[4], "|")),
		Count:      count,
	}
}

func server(filePath string, urlPath string) http.HandlerFunc {
	csvfile, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	stat, err := csvfile.Stat()
	if err != nil {
		log.Fatalln("Couldn't stat the csv file", err)
	}
	scanner := bufio.NewScanner(csvfile)
	s := make([]covince.Record, 0)

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		r := createRecordFromCsv(row)
		s = append(s, r)
	}
	log.Println(len(s), "records")

	opts := api.Opts{
		PathPrefix:  urlPath,
		MaxLineages: 16,
		GetLastModified: func() int64 {
			return stat.ModTime().UnixMilli()
		},
	}

	return api.CovinceAPI(opts, func(agg func(r covince.Record)) {
		start := time.Now()
		log.Println("Start aggregation")
		for _, r := range s {
			agg(r)
		}
		duration := time.Since(start)
		log.Println("Aggregation took:", duration.Milliseconds(), "ms")
	})
}

func serverless(filePath string) http.HandlerFunc {
	opts := api.Opts{
		MaxLineages: 16,
		GetLastModified: func() int64 {
			csvfile, err := os.Open(filePath)
			if err != nil {
				log.Fatalln("Couldn't open the csv file", err)
			}
			stat, err := csvfile.Stat()
			if err != nil {
				log.Fatalln("Couldn't stat the csv file", err)
			}
			return stat.ModTime().UnixMilli()
		},
	}

	return api.CovinceAPI(opts, func(agg func(r covince.Record)) {
		csvfile, err := os.Open(filePath)
		if err != nil {
			log.Fatalln("Couldn't open the csv file", err)
		}
		c := make(chan covince.Record, 500)
		done := make(chan bool)
		go func() {
			for r := range c {
				agg(r)
			}
			done <- true
		}()

		scanner := bufio.NewScanner(csvfile)
		for scanner.Scan() {
			row := strings.Split(scanner.Text(), ",")
			c <- createRecordFromCsv(row)
		}
		close(c)
		<-done
	})
}

func main() {
	start := time.Now()

	filePath := "aggregated.csv"
	urlPath := "/api"
	http.HandleFunc("/api/", server(filePath, urlPath))
	// http.HandleFunc("/", serverless(filePath))

	perf.LogDuration("startup", start)
	perf.LogMemory()

	http.ListenAndServe(":4000", nil)
}
