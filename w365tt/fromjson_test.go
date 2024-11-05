package w365tt

import (
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

	f365 := "../_testdata/test1.json"
	return f365
}

func TestToDb(t *testing.T) {
	fjson := getJSONfile()
	fmt.Printf("\n ***** Reading %s *****\n", fjson)
	data := LoadJSON(fjson)

	//fmt.Printf("JSON struct: %#v\n", data)

	// Save as JSON
	f := strings.TrimSuffix(fjson, filepath.Ext(fjson)) + "_db.json"
	j, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(f, j, 0666); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n ***** JSON written to %s *****\n", f)
}

func TestFromJSON(t *testing.T) {
	fjson := getJSONfile()
	fmt.Printf("\n ***** Reading %s *****\n", fjson)
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
