package timetable

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrint(t *testing.T) {
	datadir, err := filepath.Abs("../data/")
	if err != nil {
		log.Fatal(err)
	}
	//typst, err := filepath.Abs("../resources/print_timetable.typ")
	//if err != nil {
	//	log.Fatal(err)
	//}
	/*
		cmd := exec.Command("typst", "compile",
			"--root", datadir,
			"--input", "ifile="+filepath.Join("..", "_out", "ptest.json"),
			filepath.Join(datadir, "resources", "print_timetable.typ"),
			filepath.Join(datadir, "..", "ptest.pdf"))
		fmt.Printf(" ::: %s\n", cmd.String())
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(string(output))
		log.Fatalln("Quit")
	*/
	fmt.Println("\n############## TestPrint")
	const defaultPath = "../_testdata/*.w365"
	/*
		abspath, err := zenity.SelectFile(
			zenity.Filename(defaultPath),
			zenity.FileFilter{
				Name:     "Waldorf-365 files",
				Patterns: []string{"*.w365"},
				CaseFold: false,
			})
	*/
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n ***** Reading %s *****\n", abspath)
	/*
		abspath, err := filepath.Abs(w365file)
		if err != nil {
			log.Fatalf("Couldn't resolve file path: %s\n", abspath)
		}
	*/
	wzdb := w365.ReadW365(abspath)

	fmt.Println("\n +++++ GetActivities +++++")
	alist, course2activities, _ := wzbase.GetActivities(&wzdb)
	fmt.Println("\n -------------------------------------------")

	fmt.Println("\n +++++ SetLessons +++++")
	scheduleNames := []string{}
	for _, lpi := range wzdb.TableMap["LESSON_PLANS"] {
		scheduleNames = append(scheduleNames,
			wzdb.GetNode(lpi).(wzbase.LessonPlan).ID)
	}
	fmt.Printf("\n ??? Schedules: %+v\n", scheduleNames)

	if len(scheduleNames) == 0 {
		log.Fatalln("\n No Schedule")
	}
	var plan_name string
	/*
		if len(scheduleNames) == 1 {
			plan_name = scheduleNames[0]
			err = zenity.Question(
				fmt.Sprintf("Print Schedule '%s'?", plan_name),
				zenity.Title("Question"),
				zenity.QuestionIcon)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			plan_name, err = zenity.ListItems(
				"Select a 'Schedule' (timetable)",
				scheduleNames...)
			if err != nil {
				log.Fatal(err)
			}
		}
	*/
	fmt.Printf("\n Schedule: %s\n", plan_name)
	wzbase.SetLessons(&wzdb, plan_name, alist, course2activities)

	lessons := PrepareData(&wzdb, alist)
	PrintClassTimetables(lessons, plan_name, datadir,
		strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Klassen.pdf")
	PrintTeacherTimetables(lessons, plan_name, datadir,
		strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Lehrer.pdf")
	PrintRoomTimetables(lessons, plan_name, datadir,
		strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume.pdf")
	PrintRoomOverview(lessons, plan_name, datadir,
		strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume-gesamt.pdf")
}
