package main

import (
	"fmt"
	"reflect"
	"testing"
)

type KeplerTest struct {
	Filename string
	Expect   map[string]string
}

func TestKeplerName(t *testing.T) {

	filenames := []KeplerTest{
		KeplerTest{
			Filename: "kplr008462852-2013098041711_llc.fits",
			Expect:   map[string]string{"id": "008462852", "campagne": "kepler", "catalog": "KIC"},
		},
		KeplerTest{
			Filename: "ktwo246516122-c12_llc.fits",
			Expect:   map[string]string{"id": "246516122", "campagne": "ktwo", "catalog": "EPIC"},
		},
	}

	for _, filename := range filenames {

		fmt.Println(getLabels(filename.Filename))

		if reflect.DeepEqual(getLabels(filename.Filename), filename.Expect) {
			//t.Errorf("Got '%v', Expect '%v'", getLabels(filename.Filename), filename.Expect)
		}
	}
}
