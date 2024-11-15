package w365tt

import (
	"W365toFET/base"
	"encoding/json"
	"io"
	"os"
	"slices"
	"strconv"
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
	db.checkDb()
	newdb.Info = base.Info(db.Info)
	if newdb.Info.MiddayBreak == nil {
		newdb.Info.MiddayBreak = []int{}
	} else {
		// Sort and check contiguity.
		mb := newdb.Info.MiddayBreak
		slices.Sort(mb)
		if mb[len(mb)-1]-mb[0] >= len(mb) {
			base.Error.Fatalln("MiddayBreak hours not contiguous")
		}
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

	if db.Constraints == nil {
		newdb.Constraints = make(map[string]any)
	} else {
		newdb.Constraints = db.Constraints
	}
}

func (db *DbTopLevel) readDays(newdb *base.DbTopLevel) {
	for _, e := range db.Days {
		newdb.Days = append(newdb.Days, &base.Day{
			Id:   e.Id,
			Tag:  e.Tag,
			Name: e.Name,
		})
	}
}

func (db *DbTopLevel) readHours(newdb *base.DbTopLevel) {
	for i, e := range db.Hours {
		if e.Tag == "" {
			e.Tag = "(" + strconv.Itoa(i+1) + ")"
		}
		newdb.Hours = append(newdb.Hours, &base.Hour{
			Id:    e.Id,
			Tag:   e.Tag,
			Name:  e.Name,
			Start: e.Start,
			End:   e.End,
		})
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
		newdb.Teachers = append(newdb.Teachers, &base.Teacher{
			Id:               e.Id,
			Tag:              e.Tag,
			Name:             e.Name,
			Firstname:        e.Firstname,
			NotAvailable:     tsl,
			MinLessonsPerDay: e.MinLessonsPerDay,
			MaxLessonsPerDay: e.MaxLessonsPerDay,
			MaxDays:          e.MaxDays,
			MaxGapsPerDay:    e.MaxGapsPerDay,
			MaxGapsPerWeek:   e.MaxGapsPerWeek,
			MaxAfternoons:    amax,
			LunchBreak:       e.LunchBreak,
		})
		db.TeacherMap[e.Id] = true
	}
}
