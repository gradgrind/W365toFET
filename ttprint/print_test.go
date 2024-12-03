package ttprint

import (
	"W365toFET/base"
	"W365toFET/readxml"
	"W365toFET/ttbase"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var inputfiles = []string{
	"../testdata/readxml/Demo1.xml",
	"../testdata/readxml/x01.xml",
}

func TestPrint(t *testing.T) {
	base.OpenLog("")
	datadir, err := filepath.Abs("../data/")
	if err != nil {
		base.Error.Fatal(err)
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

	//

	for _, fxml := range inputfiles {
		fxml, err := filepath.Abs(fxml)
		if err != nil {
			base.Error.Fatal(err)
		}
		fmt.Println("\n ++++++++++++++++++++++")
		cdata := readxml.ConvertToDb(fxml)
		fmt.Println("*** Available Schedules:")
		slist := cdata.ScheduleNames()
		for _, sname := range slist {
			fmt.Printf("  -- %s\n", sname)
		}
		sname := "Vorlage"
		if !slices.Contains(slist, sname) {
			if len(slist) != 0 {
				sname = slist[0]
			} else {
				fmt.Println(" ... stopping ...")
				continue
			}
		}
		fmt.Printf("*** Using Schedule '%s'\n", sname)
		if !cdata.ReadSchedule(sname) {
			fmt.Println(" ... failed ...")
			continue
		}
		stempath := strings.TrimSuffix(fxml, filepath.Ext(fxml))

		stempath = strings.TrimSuffix(stempath, "_w365")

		db := cdata.Db()
		db.PrepareDb()
		ttinfo := ttbase.MakeTtInfo(db)
		doPrinting(ttinfo, datadir, stempath)
	}
}

func doPrinting(ttinfo *ttbase.TtInfo, datadir string, stempath string) {
	ordering := orderResources(ttinfo)

	/*

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
	*/
	var plan_name string = "Test Plan"
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
	//fmt.Printf("\n Schedule: %s\n", plan_name)
	//wzbase.SetLessons(&wzdb, plan_name, alist, course2activities)

	outdir := filepath.Join(filepath.Dir(stempath), "_pdf")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Println(err)
		}
	}
	outpath := filepath.Join(outdir, filepath.Base(stempath))
	//lessons := PrepareData(&wzdb, alist)
	//PrintClassTimetables(ttinfo, plan_name, datadir,
	//	strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Klassen.pdf")
	PrintTeacherTimetables(
		ttinfo, ordering, plan_name, datadir, outpath+"_Lehrer.pdf")
	/*
		PrintRoomTimetables(lessons, plan_name, datadir,
			strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume.pdf")
		PrintRoomOverview(lessons, plan_name, datadir,
			strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume-gesamt.pdf")
	*/
}
