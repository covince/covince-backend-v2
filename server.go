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

func readData(filePath string) []covince.Record {
	csvfile, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	scanner := bufio.NewScanner(csvfile)

	s := make([]covince.Record, 0)

	for scanner.Scan() {
		row := strings.Split(scanner.Text(), ",")
		count, _ := strconv.Atoi(row[4])
		r := covince.Record{
			Date:       row[0],
			PangoClade: row[1],
			Area:       row[2],
			Mutations:  row[3],
			Count:      count,
		}
		s = append(s, r)
	}

	return s
}

func main() {
	s := readData("aggregated.csv")
	log.Println(len(s), "records")
	api.CovinceAPI(api.Opts{}, s)
	http.ListenAndServe(":4000", nil)
}
