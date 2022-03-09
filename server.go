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

type Database struct {
	Count          int
	Genes          map[string]bool
	Mutations      []covince.Mutation
	MutationLookup map[string]int
	Records        []covince.Record
	Values         []covince.Value
	ValueLookup    map[string]int
}

func indexMutations(db *Database, muts []string) []*covince.Mutation {
	ptrs := make([]*covince.Mutation, len(muts))
	for i, m := range muts {
		var j int
		var ok bool
		if j, ok = db.MutationLookup[m]; !ok {
			j = len(db.Mutations)
			db.MutationLookup[m] = j

			split := strings.Split(m, ":")
			prefix := split[0]

			if _, ok = db.Genes[prefix]; !ok {
				db.Genes[prefix] = true
			}

			db.Mutations = append(
				db.Mutations,
				covince.Mutation{
					Prefix: prefix,
					Suffix: split[1],
				},
			)
		}
		ptrs[i] = &db.Mutations[j]
	}
	return ptrs
}

func indexValue(db *Database, s string) *covince.Value {
	var i int
	var ok bool
	if i, ok = db.ValueLookup[s]; !ok {
		i = len(db.Values)
		db.ValueLookup[s] = i
		db.Values = append(db.Values, covince.Value{Value: s})
	}
	return &db.Values[i]
}

func addRecordToDatabase(db *Database, row []string) {
	count, _ := strconv.Atoi(row[5])
	db.Records = append(
		db.Records,
		covince.Record{
			Area: indexValue(db, row[0]),
			Date: indexValue(db, row[1]),
			// Lineage:    indexValue(db, row[2]),
			PangoClade: indexValue(db, row[3]),
			Mutations:  indexMutations(db, strings.Split(row[4], "|")),
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
	db := Database{
		Genes:          make(map[string]bool),
		MutationLookup: make(map[string]int),
		ValueLookup:    make(map[string]int),
	}

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		addRecordToDatabase(&db, row)
	}
	log.Println(len(db.Records), "records")
	for k := range db.MutationLookup {
		delete(db.MutationLookup, k)
	}

	opts := api.Opts{
		PathPrefix:  urlPath,
		MaxLineages: 16,
		GetLastModified: func() int64 {
			return stat.ModTime().UnixMilli()
		},
		MaxSearchResults: 32,
	}

	foreach := func(agg func(r *covince.Record)) {
		start := time.Now()
		for _, r := range db.Records {
			agg(&r)
		}
		perf.LogDuration("Aggregation", start)
	}

	return api.CovinceAPI(opts, foreach, db.Genes)
}

// func serverless(filePath string) http.HandlerFunc {
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
