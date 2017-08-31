package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	warp "github.com/PierreZ/Warp10Exporter"
)

var path = flag.String("path", "", "path for csv files")
var token = flag.String("token", "write", "write token for Warp10")
var endpoint = flag.String("endpoint", "http://localhost:8080", "full warp10 endpoint address [proto]:[endpoint]:[port]")

func main() {
	flag.Parse()

	if len(*path) == 0 {
		log.Fatal("path not set")
	}
	if len(*token) == 0 {
		log.Fatal("token not set")
	}
	if len(*endpoint) == 0 {
		log.Fatal("endpoint not set")
	}

	files, err := ioutil.ReadDir(*path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !strings.Contains(file.Name(), ".csv") {
			continue
		}
		filename := file.Name()

		labels := getLabels(filename)

		gtss, err := parseCSV(*path+"/"+filename, labels)
		if err != nil {
			log.Fatalln(err)
		}

		batch := warp.NewBatch()

		for _, gts := range gtss {
			batch.AddGTS(&gts)
		}

		statuscode := batch.Push(*endpoint, *token)
		if statuscode != http.StatusOK {
			panic(fmt.Errorf("warp respond %v, exiting", statuscode))
		}
	}
	log.Println("Done!")
}

// getLabels is getting the name of the star based on filename.
//
func getLabels(filename string) map[string]string {
	labels := make(map[string]string)
	head := strings.Split(filename, "-")[0]

	labels["campagne"] = head[0:3]
	labels["id"] = head[3:]
	switch labels["campagne"] {
	case "kepler":
		labels["catalog"] = "KIC"
	case "ktwo":
		labels["catalog"] = "EPIC"
	}
	return labels
}

func parseCSV(path string, labels warp.Labels) (map[int]warp.GTS, error) {

	gtss := make(map[int]warp.GTS)

	body, err := ioutil.ReadFile(path)
	r := csv.NewReader(bytes.NewReader(body))

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range records {

		//log.Printf("line=%+v, i=%v\n", line, i)

		for j, column := range line {
			if j == 0 {
				// First column must be TIME
				continue
			}

			if contains(line, "nan") {
				continue
			}

			//log.Printf("column=%+v, j=%v\n", column, j)

			if i == 0 {
				// Create GTS
				if len(column) == 0 {
					panic("column 0")
				}
				classname := "kepler."
				column = strings.ToLower(column)
				column = strings.Replace(column, "_", ".", 1)
				classname += column

				// New GTS
				log.Println(classname)
				gts := warp.NewGTS(classname).WithLabels(labels)
				gtss[j] = *gts
				//log.Println(classname+" GTS added at index ", j)

			} else {
				// GTS exists
				ts := parseBJD(line[0])
				value := parseValue(column)
				gts := gtss[j]
				gts.AddDatapoint(ts, value)
				gtss[j] = gts
			}

		}

	}

	return gtss, nil
}

// TODO parse BJD or at leat an approximation
func parseBJD(timestr string) time.Time {

	// Removing trailing . and number
	timestr = strings.Split(timestr, ".")[0]

	i, err := strconv.ParseInt(timestr, 10, 64)
	if err != nil {
		panic(err)
	}
	return time.Unix(i, 0)
}

func parseValue(column string) float64 {
	f, err := strconv.ParseFloat(column, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.ToLower(a) == e {
			return true
		}
	}
	return false
}
