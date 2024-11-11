package readxml

import (
	"fmt"
	"testing"
)

func getXMLfile() string {
	/*
		const defaultPath = "../_testdata/*.xml"
		f365, err := zenity.SelectFile(
			zenity.Filename(defaultPath),
			zenity.FileFilter{
				Name:     "Waldorf-365 TT-export",
				Patterns: []string{"*.xml"},
				CaseFold: false,
			})
		if err != nil {
			log.Fatal(err)
		}
	*/

	f365 := "../_testdata/x01_w365.xml"
	//f365 := "../_testdata/fms1_w365.xml"
	//f365 := "../_testdata/Demo1_w365.xml"
	return f365
}

func Test2JSON(t *testing.T) {
	fxml := getXMLfile()
	fmt.Printf("\n ***** Reading %s *****\n", fxml)
	fjson := ConvertToJSON(fxml)
	fmt.Printf("\n ***** Written to %s *****\n", fjson)
}
