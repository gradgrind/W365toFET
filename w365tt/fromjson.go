package w365tt

import (
	"W365toFET/logging"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	dbdata := ReadJSON(jsonpath)
	// Days need no initialization.
	dbdata.readHours()
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

func (dbp *DbTopLevel) readHours() {
	for i := 0; i < len(dbp.Hours); i++ {
		n := dbp.Hours[i]
		if n.Tag == "" {
			n.Tag = fmt.Sprintf("(%d)", i+1)
		}
	}
}

func (dbp *DbTopLevel) readTeachers() {
	for i := 0; i < len(dbp.Teachers); i++ {
		n := dbp.Teachers[i]
		if len(n.NotAvailable) == 0 {
			// Avoid a null value
			n.NotAvailable = []TimeSlot{}
		}
		// MaxAfternoons = 0 has a special meaning (all blocked)
		if n.MaxAfternoons == 0 {
			n.MaxAfternoons = -1
			dbp.handleZeroAfternoons(&n.NotAvailable)
		}

	}
}
