package fet

import (
	"W365toFET/w365tt"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//func TestDays(t *testing.T) {
//	readDays()
//}

func TestX(t *testing.T) {
	var i interface{}
	i = "Hello"
	fmt.Printf(" --- %#v\n", getString(i))
	i = 4.5
	fmt.Printf(" --- %#v\n", getString(i))
	j := []byte("{\"a\": 1, \"b\": \"Goodbye\"}")
	json.Unmarshal(j, &i)
	fmt.Printf(" +++ %#v\n", i)
	fmt.Printf(" --- %#v\n", getString(i))
	i = w365tt.Ref("XXX1")
	fmt.Printf(" +++ %#v\n", i)
	fmt.Printf(" --- %#v\n", getString(i))
	panic("#BUG#")
}

func TestFet(t *testing.T) {
	// w365file := "../_testdata/fms.w365"
	w365file := "../_testdata/test1.json"
	abspath, err := filepath.Abs(w365file)
	if err != nil {
		log.Fatalf("Couldn't resolve file path: %s\n", abspath)
	}
	data := w365tt.LoadJSON(abspath)

	/*
			fmt.Println("\n +++++ GetActivities +++++")
			alist, course2activities, subject_courses := wzbase.GetActivities(&wzdb)
			fmt.Println("\n -------------------------------------------")
			/*
				for _, a := range alist {
					fmt.Printf(" >>> %+v\n", a)
				}

		fmt.Println("\n +++++ SetLessons +++++")
		wzbase.SetLessons(&wzdb, "Vorlage", alist, course2activities)
		/*
			for _, a := range alist {
				fmt.Printf("+++ %+v\n", a)
			}

		fmt.Println("\n +++++ SubjectActivities +++++")
		sgalist := wzbase.SubjectActivities(&wzdb,
			subject_courses, course2activities)

		/*
			type sgapr struct {
				subject    string
				groups     string
				activities []int
			}
			sgaprl := []sgapr{}
			for _, sga := range sgalist {
				s := wzdb.GetNode(sga.Subject).(wzbase.Subject).ID
				gg := []string{}
				for _, cg := range sga.Groups {
					gg = append(gg, cg.Print(wzdb))
				}
				sgaprl = append(sgaprl,
					sgapr{s, strings.Join(gg, ","), sga.Activities})
			}
			slices.SortStableFunc(sgaprl,
				func(a, b sgapr) int {
					if n := cmp.Compare(a.groups, b.groups); n != 0 {
						return n
					}
					// If names are equal, order by age
					return cmp.Compare(a.subject, b.subject)
				})
			for _, sga := range sgaprl {
				fmt.Printf("XXX %s / %s: %+v\n", sga.groups, sga.subject, sga.activities)
			}
	*/

	// ********** Build the fet file **********
	//xmlitem := make_fet_file(&data, alist, course2activities, sgalist)
	xmlitem := MakeFetFile(data)
	//fmt.Printf("\n*** fet:\n%v\n", xmlitem)
	fetfile := strings.TrimSuffix(abspath, filepath.Ext(abspath)) + ".fet"
	f, err := os.Create(fetfile)
	if err != nil {
		log.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		log.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	log.Printf("\nFET file written to: %s\n", fetfile)
}
