package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	warp "github.com/PierreZ/Warp10Exporter"
	"github.com/howeyc/fsnotify"
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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if ev.IsCreate() {
					go push(ev.Name)
				}
			case err := <-watcher.Error:
				log.Fatalln("error:", err)
				done <- true
			}
		}
	}()

	err = watcher.Watch(*path)
	if err != nil {
		log.Fatal(err)
	}

	// Hang so program doesn't exit
	<-done

	/* ... do stuff ... */
	watcher.Close()
}

func push(filename string) {
	if !strings.Contains(filename, ".csv") {
		return
	}

	labels := getLabels(filename)
	gtss, err := parseCSV(filename, labels)
	if err != nil {
		log.Fatalln(err)
	}

	batch := warp.NewBatch()

	for _, gts := range gtss {
		batch.Register(&gts)
	}
	err = batch.Push(*endpoint, *token)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("data pushed!")
	err = os.Remove(filename)
	if err != nil {
		log.Fatalln(err)
	}
}

// getLabels is getting the name of the star based on filename.
//
func getLabels(filename string) map[string]string {
	labels := make(map[string]string)
	head := strings.Split(filename, "-")[0]

	labels["campagne"] = head[0:4]
	labels["id"] = head[4:]
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

		for j, column := range line {
			if j == 0 {
				// First column must be TIME
				continue
			}

			if contains(line, "nan") {
				continue
			}

			if i == 0 {
				// Create GTS
				if len(column) == 0 {
					panic("column 0")
				}
				classname := "kepler."
				column = strings.ToLower(column)
				column = strings.Replace(column, "_", ".", -1)
				classname += column

				// New GTS
				gts := warp.NewGTS(classname).WithLabels(labels)
				gtss[j] = *gts
				//log.Println(classname+" GTS added at index ", j)

			} else {
				// GTS exists
				ts := parseBJD(line[0])

				value := parseScientificValue(column)
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

	return time.Unix(int64(parseScientificValue(timestr)), 0)
}

func parseScientificValue(s string) (f float64) {

	split := strings.Split(s, "e")
	f, err := strconv.ParseFloat(split[0], 64)
	if err != nil {
		panic(err)
	}
	// handling "e(+|-)"
	if len(split) == 2 {
		exp, _ := strconv.ParseFloat(split[1][1:], 64)
		if split[1][0:1] == "-" {
			f *= -1
		}
		for j := 0; j < int(exp); j++ {
			f *= 10
		}
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
