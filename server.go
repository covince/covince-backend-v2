package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/covince/covince-backend-v2/api"
	"github.com/covince/covince-backend-v2/covince"
)

func createRecordFromCsv(row []string) covince.Record {
	count, _ := strconv.Atoi(row[4])
	return covince.Record{
		Date:       row[0],
		PangoClade: row[1],
		Area:       row[2],
		Mutations:  row[3],
		Count:      count,
	}
}

func server(filePath string) {
	csvfile, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	scanner := bufio.NewScanner(csvfile)
	s := make([]covince.Record, 0)

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		r := createRecordFromCsv(row)
		s = append(s, r)
	}
	log.Println(len(s), "records")

	api.CovinceAPI(api.Opts{MaxLineages: 16}, func(agg func(r covince.Record)) {
		for _, r := range s {
			agg(r)
		}
	})
}

func serverless(filePath string) {
	api.CovinceAPI(api.Opts{MaxLineages: 16}, func(agg func(r covince.Record)) {
		csvfile, err := os.Open(filePath)
		// csvfile, err := os.Open("input2.tsv")
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
	server("aggregated.csv")
	// serverless("aggregated.csv")

	http.ListenAndServe(":4000", nil)
}
