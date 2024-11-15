package w365tt

import (
	"W365toFET/base"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getJSONfile() string {
	/*
		const defaultPath = "../_testdata/*.json"
		f365, err := zenity.SelectFile(
			zenity.Filename(defaultPath),
			zenity.FileFilter{
				Name:     "Waldorf-365 TT-export",
				Patterns: []string{"*.json"},
				CaseFold: false,
			})
		if err != nil {
			log.Fatal(err)
		}
	*/

	//f365 := "../_testdata/fms1_w365.json"
	f365 := "../_testdata/Demo1_w365.json"
	return f365
}

func TestFromJSON(t *testing.T) {
	// Test reading a _w365.json file without any processing.
	fjson := getJSONfile()
	fmt.Printf("\n ***** Reading %s *****\n", fjson)

	stempath := strings.TrimSuffix(fjson, filepath.Ext(fjson))
	logpath := stempath + ".log"
	base.OpenLog(logpath)

	data := ReadJSON(fjson)

	//fmt.Printf("JSON struct: %#v\n", data)

	// Save as JSON
	f := strings.TrimSuffix(fjson, filepath.Ext(fjson)) + "_2.json"
	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(f, j, 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n ***** JSON written to %s *****\n", f)
}

func TestToDb(t *testing.T) {
	// Test reading a _w365.json file including initial processing.
	fjson := getJSONfile()
	fmt.Printf("\n ***** Reading %s *****\n", fjson)

	stempath := strings.TrimSuffix(fjson, filepath.Ext(fjson))
	logpath := stempath + ".log"
	base.OpenLog(logpath)

	db := base.NewDb()
	LoadJSON(db, fjson)
	db.PrepareDb()

	// Save as JSON
	stempath = strings.TrimSuffix(stempath, "_w365")
	f := stempath + "_db.json"
	j, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(f, j, 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n ***** JSON written to %s *****\n", f)
}
