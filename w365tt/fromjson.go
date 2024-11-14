package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
	"encoding/json"
	"io"
	"os"
	"strconv"
)

// Read to the local, tweaked DbTopLevel
func ReadJSON(jsonpath string) *DbTopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(jsonpath)
	if err != nil {
		logging.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	logging.Message.Printf("*+ Reading: %s\n", jsonpath)
	v := DbTopLevel{}
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
		logging.Error.Fatalf("Could not unmarshal json: %s\n", err)
	}
	return &v
}

func LoadJSON(jsonpath string) *DbTopLevel {
	db := ReadJSON(jsonpath)
	newdb := &base.DbTopLevel{}
	newdb.Info = base.Info(db.Info)
	db.readDays(newdb)
	db.readHours(newdb)
	db.readTeachers(newdb)
	db.readSubjects(newdb)
	db.readRooms(newdb)

	dbdata.checkDb()
	dbdata.readTeachers()
	dbdata.readSubjects()
	dbdata.readRooms()
	dbdata.readRoomGroups()
	dbdata.readRoomChoiceGroups()
	// W365 has no RoomChoicesGroups: – they must be generated from the
	// PreferredRooms lists of courses.
	// To manage potentially incomplete Tag and Name fields for RoomGroups
	// from W365, perform the checking after all room types have been "read".
	dbdata.checkRoomGroups()
	dbdata.readClasses()
	dbdata.readCourses()
	dbdata.readSuperCourses()
	dbdata.readSubCourses()
	dbdata.readLessons()
	return dbdata
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
	}
}
