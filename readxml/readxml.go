package readxml

import (
	"W365toFET/base"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

const SCHEDULE_NAME = "Vorlage"

type Ref = base.Ref // Element reference

type conversionData struct {
	db          *base.DbTopLevel
	xmlin       *Scenario
	categories  map[Ref]*Category
	absences    map[Ref]base.TimeSlot
	divisions   map[Ref]*Division
	subjectMap  map[Ref]*base.Subject
	subjectTags map[string]Ref
}

func newConversionData(xmlin *Scenario) *conversionData {
	return &conversionData{
		db:          base.NewDb(),
		xmlin:       xmlin,
		categories:  map[Ref]*Category{},
		absences:    map[Ref]base.TimeSlot{},
		divisions:   map[Ref]*Division{},
		subjectMap:  map[Ref]*base.Subject{},
		subjectTags: map[string]Ref{},
	}
}

func ReadXML(xmlpath string) W365XML {
	// Open the  XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(xmlFile)
	base.Message.Printf("Reading: %s\n", xmlpath)
	v := W365XML{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		base.Error.Fatalf("XML error in %s:\n %v\n", xmlpath, err)
	}
	return v
}

func ConvertToDb(f365xml string) *base.DbTopLevel {
	root := ReadXML(f365xml)
	a := root.SchoolState.ActiveScenario
	var indata *Scenario
	for i := 0; i < len(root.Scenarios); i++ {
		sp := &root.Scenarios[i]
		if sp.Id == a {
			indata = sp
			break
		}
	}
	if indata == nil {
		base.Error.Fatalln("No Active Scenario")
	}

	cdata := newConversionData(indata)
	db := cdata.db
	db.Info.Reference = string(indata.Id)
	db.Info.Institution = root.SchoolState.SchoolName
	//db.Info.Schedule = "Vorlage"

	// These items don't correspond directly to base db elements:
	cdata.readCategories()
	cdata.readAbsences()
	cdata.readDivisions()

	// db elements
	cdata.readDays()
	cdata.readHours()
	cdata.readSubjects()
	cdata.readRooms()
	cdata.readTeachers()
	cdata.readGroups()
	cdata.readDivisions()
	cdata.readClasses()

	courseLessons := readCourses(db, id2node, indata.Courses)
	// courseLessons maps course ref -> list of lesson lengths
	readLessons(id2node, indata.Lessons)
	schedmap := readSchedules(id2node, indata.Schedules)

	llist, ok := schedmap[SCHEDULE_NAME]
	if !ok {
		base.Warning.Printf("No Schedule with Name=%s\n", SCHEDULE_NAME)
	}

	// Generate Lessons
	makeLessons(db, id2node, courseLessons, llist)

	return cdata.db
}

func (cdata *conversionData) readCategories() {
	for i := 0; i < len(cdata.xmlin.Categories); i++ {
		n := &cdata.xmlin.Categories[i]
		cdata.categories[n.Id] = n
	}
}

func (cdata *conversionData) readAbsences() {
	for i := 0; i < len(cdata.xmlin.Absences); i++ {
		n := &cdata.xmlin.Absences[i]
		e := base.TimeSlot{
			Day:  n.Day,
			Hour: n.Hour,
		}
		cdata.absences[n.Id] = e
	}
}

func (cdata *conversionData) getAbsences(
	reflist RefList,
	msg string,
) []base.TimeSlot {
	result := []base.TimeSlot{}
	for _, aref := range splitRefList(reflist) {
		ts, ok := cdata.absences[aref]
		if !ok {
			base.Error.Fatalf("%s:\n  -- Invalid Absence: %s\n", msg, aref)
		}
		result = append(result, ts)
	}
	slices.SortFunc(result, func(a, b base.TimeSlot) int {
		if a.Day < b.Day {
			return -1
		}
		if a.Day == b.Day {
			if a.Hour < b.Hour {
				return -1
			}
			if a.Hour == b.Hour {
				base.Error.Fatalf("%s:\n  -- Equal Absences\n", msg)
			}
			return 1
		}
		return 1
	})
	return result
}

func (cdata *conversionData) readDays() {
	slices.SortFunc(cdata.xmlin.Days, func(a, b Day) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	for i := 0; i < len(cdata.xmlin.Days); i++ {
		n := &cdata.xmlin.Days[i]
		e := cdata.db.NewDay(n.Id)
		e.Name = n.Name
		e.Tag = n.Shortcut
	}
}

func (cdata *conversionData) readHours() {
	slices.SortFunc(cdata.xmlin.Hours, func(a, b Hour) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	db := cdata.db
	for i := 0; i < len(cdata.xmlin.Hours); i++ {
		n := &cdata.xmlin.Hours[i]
		e := cdata.db.NewHour(n.Id)
		e.Name = n.Name
		e.Tag = n.Shortcut

		t0 := get_time(n.Start)
		t1 := get_time(n.End)
		if len(t0) != 0 {
			e.Start = t0
			e.End = t1
		}

		if n.FirstAfternoonHour {
			db.Info.FirstAfternoonHour = i
		}
		if n.MiddayBreak {
			db.Info.MiddayBreak = append(
				db.Info.MiddayBreak, i)
		}
	}
}

func get_time(t string) string {
	// Check time and return as "mm:hh"
	tn := strings.Split(t, ":")
	if len(tn) < 2 {
		return ""
	}
	h, err := strconv.Atoi(tn[0])
	if err != nil || h > 23 || h < 0 {
		return ""
	}
	m, err := strconv.Atoi(tn[1])
	if err != nil || m > 59 || m < 0 {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", h, m)
}

func readSchedules(
	id2node map[Ref]interface{},
	items []Schedule,
) map[string][]Ref {
	// These serve only to determine which Lesson elements are relevant.
	smap := map[string][]Ref{} // Name -> Lesson Ref list
	for _, n := range items {
		msg := fmt.Sprintf("Bad Lesson Ref in Schedule %s", n.Id)
		smap[n.Name] = GetRefList(id2node, n.Lessons, msg)
	}
	return smap
}

func splitRefList(reflist RefList) []Ref {
	result := []Ref{}
	for _, ref := range strings.Split(string(reflist), ",") {
		result = append(result, Ref(ref))
	}
	return result
}

// Block all afternoons if nAfternnons == 0.
func handleZeroAfternoons(
	dbp *base.DbTopLevel,
	notAvailable []base.TimeSlot,
	nAfternoons int,
) []base.TimeSlot {
	if nAfternoons != 0 {
		return notAvailable
	}
	// Make a bool array and fill this in two passes, then remake list.
	// NOTE: Days, Hours and Info must already be set up in the base db.
	namap := make([][]bool, len(dbp.Days))
	nhours := len(dbp.Hours)
	// In the first pass, block afternoons
	for i := range namap {
		namap[i] = make([]bool, nhours)
		for h := dbp.Info.FirstAfternoonHour; h < nhours; h++ {
			namap[i][h] = true
		}
	}
	// In the second pass, include existing blocked hours.
	for _, ts := range notAvailable {
		namap[ts.Day][ts.Hour] = true
	}
	// Build a new base.TimeSlot list
	na := []base.TimeSlot{}
	for d, naday := range namap {
		for h, nahour := range naday {
			if nahour {
				na = append(na, base.TimeSlot{Day: d, Hour: h})
			}
		}
	}
	return na
}
