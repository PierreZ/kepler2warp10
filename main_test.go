package main

import "testing"

type KeplerTest struct {
	Filename string
	Expect   string
}

func TestKeplerName(t *testing.T) {

	filenames := []KeplerTest{
		KeplerTest{
			Filename: "kplr008462852-2013098041711_llc.fits",
			Expect:   "008462852",
		},
		KeplerTest{
			Filename: "ktwo246516122-c12_llc.fits",
			Expect:   "246516122",
		},
	}

	for _, filename := range filenames {
		if getEPICIDFromFilename(filename.Filename) != filename.Expect {
			t.Errorf("Got '%v', Expect '%v'", getEPICIDFromFilename(filename.Filename), filename.Expect)
		}
	}
}
