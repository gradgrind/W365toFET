package w365tt

import (
	"W365toFET/base"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
)

// Read to the local, tweaked DbTopLevel
func ReadJSON(jsonpath string) *DbTopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(jsonpath)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	base.Message.Printf("*+ Reading: %s\n", jsonpath)
	v := DbTopLevel{}
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
		base.Error.Fatalf("Could not unmarshal json: %s\n", err)
	}
	return &v
}

func LoadJSON(newdb *base.DbTopLevel, jsonpath string) {
	db := ReadJSON(jsonpath)
	newdb.Info = base.Info(db.Info)
	newdb.ModuleData = map[string]any{
		"PrintTables": db.PrintTables,
		"FetData":     db.FetData,
	}
	db.readDays(newdb)
	db.readHours(newdb)
	db.readTeachers(newdb)
	db.readSubjects(newdb)
	db.readRooms(newdb)
	db.readRoomGroups(newdb)
	// To manage potentially incomplete Tag and Name fields for RoomGroups
	// from W365, perform the checking after all room types have been "read".
	db.checkRoomGroups(newdb)
	db.readClasses(newdb)
	db.readCourses(newdb)
	db.readSuperCourses(newdb)
	db.readLessons(newdb)
	db.readConstraints(newdb)
}

func (db *DbTopLevel) readDays(newdb *base.DbTopLevel) {
	for _, e := range db.Days {
		n := newdb.NewDay(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
	}
}

func (db *DbTopLevel) readHours(newdb *base.DbTopLevel) {
	for i, e := range db.Hours {
		tag := e.Tag
		if tag == "" {
			tag = "(" + strconv.Itoa(i+1) + ")"
		}
		n := newdb.NewHour(e.Id)
		n.Tag = tag
		n.Name = e.Name
		// If the input times have seconds, strip these off.
		ts := strings.Split(e.Start, ":")
		if len(ts) == 3 {
			n.Start = ts[0] + ":" + ts[1]
		} else {
			n.Start = e.Start
		}
		ts = strings.Split(e.End, ":")
		if len(ts) == 3 {
			n.End = ts[0] + ":" + ts[1]
		} else {
			n.End = e.End
		}
	}
}

func (db *DbTopLevel) readTeachers(newdb *base.DbTopLevel) {
	db.TeacherMap = map[base.Ref]bool{}
	for _, e := range db.Teachers {
		// MaxAfternoons = 0 has a special meaning (all blocked)
		amax := e.MaxAfternoons
		tsl := db.handleZeroAfternoons(e.NotAvailable, amax)
		if amax == 0 {
			amax = -1
		}
		n := newdb.NewTeacher(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
		n.Firstname = e.Firstname
		n.NotAvailable = tsl
		n.MinLessonsPerDay = e.MinLessonsPerDay
		n.MaxLessonsPerDay = e.MaxLessonsPerDay
		n.MaxDays = e.MaxDays
		n.MaxGapsPerDay = e.MaxGapsPerDay
		n.MaxGapsPerWeek = e.MaxGapsPerWeek
		n.MaxAfternoons = amax
		n.LunchBreak = e.LunchBreak

		db.TeacherMap[e.Id] = true
	}
}
