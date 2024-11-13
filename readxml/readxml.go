package readxml

import (
	"W365toFET/w365tt"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const SCHEDULE_NAME = "Vorlage"

func ReadXML(xmlpath string) W365XML {
	// Open the  XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		log.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(xmlFile)
	log.Printf("*+ Reading: %s\n", xmlpath)
	v := W365XML{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		log.Fatalf("XML error in %s:\n %v\n", xmlpath, err)
	}
	v.Path = xmlpath
	return v
}

func ConvertToJSON(f365xml string) string {
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
		log.Fatalln("*ERROR* No Active Scenario")
	}
	outdata := w365tt.DbTopLevel{}
	id2node := map[w365tt.Ref]interface{}{}

	outdata.Info.Reference = string(indata.Id)
	outdata.Info.Institution = root.SchoolState.SchoolName
	//outdata.Info.Schedule = "Vorlage"
	readCategories(id2node, indata.Categories)
	readDays(&outdata, id2node, indata.Days)
	readHours(&outdata, id2node, indata.Hours)
	for _, n := range indata.Absences {
		id2node[n.IdStr()] = n
	}
	readSubjects(&outdata, id2node, indata.Subjects)
	readRooms(&outdata, id2node, indata.Rooms)
	readTeachers(&outdata, id2node, indata.Teachers)
	readGroups(&outdata, id2node, indata.Groups)
	for _, n := range indata.Divisions {
		id2node[n.IdStr()] = n
	}
	readClasses(&outdata, id2node, indata.Classes)
	courseLessons := readCourses(&outdata, id2node, indata.Courses)
	// courseLessons maps course ref -> list of lesson lengths
	readLessons(id2node, indata.Lessons)
	schedmap := readSchedules(id2node, indata.Schedules)

	llist, ok := schedmap[SCHEDULE_NAME]
	if !ok {
		log.Printf("*WARNING* No Schedule with Name=%s\n", SCHEDULE_NAME)
	}

	// Generate w365tt.Lessons
	makeLessons(&outdata, id2node, courseLessons, llist)

	// Save as JSON
	f := strings.TrimSuffix(root.Path, filepath.Ext(root.Path)) + ".json"
	j, err := json.MarshalIndent(outdata, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(f, j, 0666); err != nil {
		log.Fatal(err)
	}
	return f
}

func addId(id2node map[w365tt.Ref]interface{}, node TTNode) w365tt.Ref {
	// Check for redeclarations
	nid := node.IdStr()
	if _, ok := id2node[nid]; ok {
		log.Printf("Redefinition of %s\n", nid)
		return ""
	}
	id2node[nid] = node
	return nid
}

func readDays(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Day,
) {
	slices.SortFunc(items, func(a, b Day) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		outdata.Days = append(outdata.Days, w365tt.Day{
			Id:   nid,
			Name: n.Name,
			Tag:  n.Shortcut,
		})
	}
}

func readHours(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Hour,
) {
	slices.SortFunc(items, func(a, b Hour) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	for i, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		r := w365tt.Hour{
			Id:   nid,
			Name: n.Name,
			Tag:  n.Shortcut,
		}
		t0 := get_time(n.Start)
		t1 := get_time(n.End)
		if len(t0) != 0 {
			r.Start = t0
			r.End = t1
		}
		outdata.Hours = append(outdata.Hours, r)

		if n.FirstAfternoonHour {
			outdata.Info.FirstAfternoonHour = i
		}
		if n.MiddayBreak {
			outdata.Info.MiddayBreak = append(
				outdata.Info.MiddayBreak, i)
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

func readSubjects(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Subject,
) {
	slices.SortFunc(items, func(a, b Subject) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		outdata.Subjects = append(outdata.Subjects, w365tt.Subject{
			Id:   nid,
			Name: n.Name,
			Tag:  n.Shortcut,
		})
	}
}

func readRooms(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Room,
) {
	slices.SortFunc(items, func(a, b Room) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	rglist := map[w365tt.Ref]Room{} // RoomGroup elements
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		// Extract RoomGroup elements
		if n.RoomGroups != "" {
			rglist[nid] = n
			continue
		}
		// Normal Room
		r := w365tt.Room{
			Id:   nid,
			Name: n.Name,
			Tag:  n.Shortcut,
		}
		msg := fmt.Sprintf("Room %s in Absences", nid)
		for _, ai := range GetRefList(id2node, n.Absences, msg) {
			an := id2node[ai].(Absence)
			r.NotAvailable = append(r.NotAvailable, w365tt.TimeSlot{
				Day:  an.Day,
				Hour: an.Hour,
			})
		}
		sortAbsences(r.NotAvailable)
		outdata.Rooms = append(outdata.Rooms, r)
	}
	// Now handle the RoomGroups
	for nid, n := range rglist {
		msg := fmt.Sprintf("Room %s in RoomGroups", nid)
		rg := GetRefList(id2node, n.RoomGroups, msg)
		r := w365tt.RoomGroup{
			Id:   nid,
			Name: n.Shortcut, // !
			//Tag: n.Shortcut,
			Rooms: rg,
		}
		outdata.RoomGroups = append(outdata.RoomGroups, r)
	}
}

func readTeachers(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Teacher,
) {
	slices.SortFunc(items, func(a, b Teacher) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(outdata.Days)
	nhours := len(outdata.Hours)
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
		r := w365tt.Teacher{
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
			r.NotAvailable = append(r.NotAvailable, w365tt.TimeSlot{
				Day:  an.(Absence).Day,
				Hour: an.(Absence).Hour,
			})
		}
		sortAbsences(r.NotAvailable)
		outdata.Teachers = append(outdata.Teachers, r)
	}
}

func sortAbsences(alist []w365tt.TimeSlot) {
	slices.SortFunc(alist, func(a, b w365tt.TimeSlot) int {
		if a.Day < b.Day {
			return -1
		}
		if a.Day == b.Day {
			if a.Hour < b.Hour {
				return -1
			}
			if a.Hour == b.Hour {
				log.Fatalln("Equal Absences")
			}
			return 1
		}
		return 1
	})
}

func readGroups(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Group,
) {
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		outdata.Groups = append(outdata.Groups, w365tt.Group{
			Id:  nid,
			Tag: n.Shortcut,
		})
	}
}

func readClasses(
	outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Class,
) {
	slices.SortFunc(items, func(a, b Class) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(outdata.Days)
	nhours := len(outdata.Hours)
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
		maxpm := n.MaxAfternoons
		if maxpm >= ndays {
			maxpm = -1
		}
		var r w365tt.Class
		if isStandIns(id2node, n.Categories, nid) {
			r = w365tt.Class{
				Id:               nid,
				Name:             n.Name,
				Year:             -1,
				Letter:           "",
				Tag:              "",
				Divisions:        []w365tt.Division{},
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
			r = w365tt.Class{
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
			r.Divisions = []w365tt.Division{}
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
					r.Divisions = append(r.Divisions, w365tt.Division{
						Name:   nm,
						Groups: glist,
					})
				}
			}
		}
		msg := fmt.Sprintf("Class %s in Absences", nid)
		for _, ai := range GetRefList(id2node, n.Absences, msg) {
			an := id2node[ai]
			r.NotAvailable = append(r.NotAvailable, w365tt.TimeSlot{
				Day:  an.(Absence).Day,
				Hour: an.(Absence).Hour,
			})
		}
		sortAbsences(r.NotAvailable)
		outdata.Classes = append(outdata.Classes, r)
	}
}

func readSchedules(
	id2node map[w365tt.Ref]interface{},
	items []Schedule,
) map[string][]w365tt.Ref {
	// These serve only to determine which Lesson elements are relevant.
	smap := map[string][]w365tt.Ref{} // Name -> Lesson Ref list
	for _, n := range items {
		msg := fmt.Sprintf("Bad Lesson Ref in Schedule %s", n.Id)
		smap[n.Name] = GetRefList(id2node, n.Lessons, msg)
	}
	return smap
}

func readCategories(
	//outdata *w365tt.DbTopLevel,
	id2node map[w365tt.Ref]interface{},
	items []Category,
) {
	for _, n := range items {
		nid := addId(id2node, &n)
		if nid == "" {
			continue
		}
	}
}

func GetRefList(
	id2node map[w365tt.Ref]interface{},
	reflist RefList,
	messages ...string,
) []w365tt.Ref {
	var rl []w365tt.Ref
	if reflist != "" {
		for _, rs := range strings.Split(string(reflist), ",") {
			rr := w365tt.Ref(rs)
			if _, ok := id2node[rr]; ok {
				rl = append(rl, rr)
			} else {
				log.Printf("Invalid Reference in RefList: %s\n", rs)
				for _, msg := range messages {
					log.Printf("  ++ %s\n", msg)
				}
			}
		}
	}
	return rl
}
