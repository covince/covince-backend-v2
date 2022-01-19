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
)

func createRecordFromCsv(row []string) covince.Record {
	count, _ := strconv.Atoi(row[5])
	return covince.Record{
		Date:       row[0],
		Lineage:    row[1],
		PangoClade: row[2],
		Area:       row[3],
		Mutations:  row[4],
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
	filePath := "aggregated.csv"
	urlPath := "/api/raw"
	http.HandleFunc("/api/raw/", server(filePath, urlPath))
	// http.HandleFunc("/", serverless(filePath))

	http.ListenAndServe(":4000", nil)
}
