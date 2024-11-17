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
	SubjectMap  map[Ref]*base.Subject
	SubjectTags map[string]Ref
}

func newConversionData(xmlin *Scenario) *conversionData {
	return &conversionData{
		db:          base.NewDb(),
		xmlin:       xmlin,
		categories:  map[Ref]*Category{},
		absences:    map[Ref]base.TimeSlot{},
		SubjectMap:  map[Ref]*base.Subject{},
		SubjectTags: map[string]Ref{},
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
	cdata.readCategories()
	cdata.readAbsences()
	cdata.readDays()
	cdata.readHours()
	cdata.readSubjects()
	readRooms(db, id2node, indata.Rooms)
	readTeachers(db, id2node, indata.Teachers)
	readGroups(db, id2node, indata.Groups)
	for _, n := range indata.Divisions {
		id2node[n.IdStr()] = n
	}
	readClasses(db, id2node, indata.Classes)
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
	for _, aref := range SplitRefList(reflist) {
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

func readTeachers(
	outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	items []Teacher,
) {
	slices.SortFunc(items, func(a, b Teacher) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(db.Days)
	nhours := len(db.Hours)
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		maxdays := n.MaxDays
		if maxdays >= ndays {
			maxdays = -1
		}
		maxpm := n.MaxAfternoons
		if maxpm >= ndays {
			maxpm = -1
		}
		lb := withLunchBreak(id2node, n.Categories, nid)
		maxlpd := n.MaxLessonsPerDay
		if lb {
			if maxlpd >= nhours-1 {
				maxlpd = -1
			}
		} else if maxlpd >= nhours {
			maxlpd = -1
		}
		r := &base.Teacher{
			Id:               nid,
			Name:             n.Name,
			Tag:              n.Shortcut,
			Firstname:        n.Firstname,
			MinLessonsPerDay: n.MinLessonsPerDay,
			MaxLessonsPerDay: maxlpd,
			MaxDays:          maxdays,
			MaxGapsPerDay:    n.MaxGapsPerDay,
			MaxGapsPerWeek:   -1,
			MaxAfternoons:    maxpm,
			LunchBreak:       lb,
		}
		msg := fmt.Sprintf("Teacher %s in Absences", nid)
		for _, ai := range GetRefList(id2node, n.Absences, msg) {
			an := id2node[ai]
			r.NotAvailable = append(r.NotAvailable, base.TimeSlot{
				Day:  an.(Absence).Day,
				Hour: an.(Absence).Hour,
			})
		}
		sortAbsences(r.NotAvailable)
		db.Teachers = append(db.Teachers, r)
	}
}

func readGroups(
	outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	items []Group,
) {
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		db.Groups = append(db.Groups, &base.Group{
			Id:  nid,
			Tag: n.Shortcut,
		})
	}
}

func readClasses(
	outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	items []Class,
) {
	slices.SortFunc(items, func(a, b Class) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(db.Days)
	nhours := len(db.Hours)
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		maxpm := n.MaxAfternoons
		if maxpm >= ndays {
			maxpm = -1
		}
		var r *base.Class
		if isStandIns(id2node, n.Categories, nid) {
			r = &base.Class{
				Id:               nid,
				Name:             n.Name,
				Year:             -1,
				Letter:           "",
				Tag:              "",
				Divisions:        []base.Division{},
				MinLessonsPerDay: -1,
				MaxLessonsPerDay: -1,
				MaxGapsPerDay:    -1,
				MaxGapsPerWeek:   -1,
				MaxAfternoons:    -1,
				LunchBreak:       false,
				ForceFirstHour:   false,
			}
		} else {
			lb := withLunchBreak(id2node, n.Categories, nid)
			maxlpd := n.MaxLessonsPerDay
			if lb {
				if maxlpd >= nhours-1 {
					maxlpd = -1
				}
			} else if maxlpd >= nhours {
				maxlpd = -1
			}
			r = &base.Class{
				Id:               nid,
				Name:             n.Name,
				Year:             n.Level,
				Letter:           n.Letter,
				Tag:              fmt.Sprintf("%d%s", n.Level, n.Letter),
				MinLessonsPerDay: n.MinLessonsPerDay,
				MaxLessonsPerDay: maxlpd,
				MaxGapsPerDay:    -1,
				MaxGapsPerWeek:   0,
				MaxAfternoons:    maxpm,
				LunchBreak:       lb,
				ForceFirstHour:   n.ForceFirstHour,
			}
			// Initialize Divisions to get [] instead of null, when empty
			r.Divisions = []base.Division{}
			msg := fmt.Sprintf("Class %s in Divisions", nid)
			for i, d := range GetRefList(id2node, n.Divisions, msg) {
				dn := id2node[d].(Division)
				msg = fmt.Sprintf("Division %s in Groups", d)
				glist := GetRefList(id2node, dn.Groups, msg)
				if len(glist) != 0 {
					nm := dn.Name
					if nm == "" {
						nm = fmt.Sprintf("#div%d", i)
					}
					r.Divisions = append(r.Divisions, base.Division{
						Name:   nm,
						Groups: glist,
					})
				}
			}
		}
		msg := fmt.Sprintf("Class %s in Absences", nid)
		for _, ai := range GetRefList(id2node, n.Absences, msg) {
			an := id2node[ai]
			r.NotAvailable = append(r.NotAvailable, base.TimeSlot{
				Day:  an.(Absence).Day,
				Hour: an.(Absence).Hour,
			})
		}
		sortAbsences(r.NotAvailable)
		db.Classes = append(db.Classes, r)
	}
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

func SplitRefList(reflist RefList) []Ref {
	result := []Ref{}
	for _, ref := range strings.Split(string(reflist), ",") {
		result = append(result, Ref(ref))
	}
	return result
}

// TODO--? If I am not maintaining id2node, this won't work completely ...
// Categories are separate.
func GetRefList(
	id2node map[Ref]interface{},
	reflist RefList,
	messages ...string,
) []Ref {
	var rl []Ref
	if reflist != "" {
		for _, rs := range strings.Split(string(reflist), ",") {
			rr := Ref(rs)
			if _, ok := id2node[rr]; ok {
				rl = append(rl, rr)
			} else {
				msglist := []string{
					fmt.Sprintf("Invalid Reference in RefList: %s\n", rs)}
				for _, msg := range messages {
					msglist = append(msglist, fmt.Sprintf("  ++ %s\n", msg))
				}
				base.Error.Printf(strings.Join(msglist, ""))
			}
		}
	}
	return rl
}
