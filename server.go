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

func addRecordToDatabase(db *covince.Database, row []string) {
	count, _ := strconv.Atoi(row[5])
	db.Records = append(
		db.Records,
		covince.Record{
			Area: db.IndexValue(row[0]),
			Date: db.IndexValue(row[1]),
			// Lineage:    db.IndexValue(row[2]),
			PangoClade: db.IndexValue(row[3]),
			Mutations:  db.IndexMutations(strings.Split(row[4], "|"), ":"),
			Count:      count,
		},
	)
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
	db := covince.CreateDatabase()

	buf := []byte{}
	// increase the buffer size to 2Mb
	scanner.Buffer(buf, 2048*1024)

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		addRecordToDatabase(db, row)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("%v", err)
	}

	log.Println(len(db.Records), "records")
	for k := range db.MutationLookup {
		delete(db.MutationLookup, k)
	}

	opts := api.Opts{
		PathPrefix:       urlPath,
		MaxLineages:      16,
		Genes:            db.Genes,
		MaxSearchResults: 32,
		LastModified:     stat.ModTime().UnixMilli(),
	}

	foreach := func(agg func(r *covince.Record), sliceIndex int) {
		start := time.Now()
		for _, r := range db.Records {
			agg(&r)
		}
		perf.LogDuration("Aggregation", start)
	}

	return api.CovinceAPI(opts, foreach)
}

// func serverless(filePath string) http.HandlerFunc {

//	TODO: update example to read opts from JSON
// 	opts := api.Opts{
// 		MaxLineages: 16,
// 		GetLastModified: func() int64 {
// 			csvfile, err := os.Open(filePath)
// 			if err != nil {
// 				log.Fatalln("Couldn't open the csv file", err)
// 			}
// 			stat, err := csvfile.Stat()
// 			if err != nil {
// 				log.Fatalln("Couldn't stat the csv file", err)
// 			}
// 			return stat.ModTime().UnixMilli()
// 		},
// 	}

// 	return api.CovinceAPI(opts, func(agg func(r covince.Record)) {
// 		csvfile, err := os.Open(filePath)
// 		if err != nil {
// 			log.Fatalln("Couldn't open the csv file", err)
// 		}
// 		c := make(chan covince.Record, 500)
// 		done := make(chan bool)
// 		go func() {
// 			for r := range c {
// 				agg(r)
// 			}
// 			done <- true
// 		}()

// 		scanner := bufio.NewScanner(csvfile)
// 		for scanner.Scan() {
// 			row := strings.Split(scanner.Text(), ",")
// 			c <- createRecordFromCsv(row)
// 		}
// 		close(c)
// 		<-done
// 	})
// }

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
