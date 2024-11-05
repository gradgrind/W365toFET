package w365tt

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// Read to the local, tweaked DbTopLevel
func ReadJSON(jsonpath string) *DbTopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(jsonpath)
	if err != nil {
		log.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	log.Printf("*+ Reading: %s\n", jsonpath)
	v := DbTopLevel{}
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
		log.Fatalf("Could not unmarshal json: %s\n", err)
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
	mdbok := len(dbp.Info.MiddayBreak) == 0
	for i := 0; i < len(dbp.Hours); i++ {
		n := &dbp.Hours[i]
		if n.FirstAfternoonHour {
			dbp.Info.FirstAfternoonHour = i
			n.FirstAfternoonHour = false
		}
		if n.MiddayBreak {
			if mdbok {
				dbp.Info.MiddayBreak = append(
					dbp.Info.MiddayBreak, i)
			} else {
				log.Println("*ERROR* MiddayBreak set in Info AND Hours")
			}
			n.MiddayBreak = false
		}
		if n.Tag == "" {
			n.Tag = fmt.Sprintf("(%d)", i+1)
		}
	}
}

func (dbp *DbTopLevel) readTeachers() {
	for i := 0; i < len(dbp.Teachers); i++ {
		n := &dbp.Teachers[i]
		if len(n.NotAvailable) == 0 {
			// Avoid a null value
			n.NotAvailable = []TimeSlot{}
		}
		if n.MinLessonsPerDay == nil {
			n.MinLessonsPerDay = -1.0
		}
		if n.MaxLessonsPerDay == nil {
			n.MaxLessonsPerDay = -1.0
		}
		if n.MaxGapsPerDay == nil {
			n.MaxGapsPerDay = -1.0
		}
		if n.MaxGapsPerWeek == nil {
			n.MaxGapsPerWeek = -1.0
		}
		if n.MaxDays == nil {
			n.MaxDays = -1.0
		}
		if n.MaxAfternoons == nil {
			n.MaxAfternoons = -1.0
		}
	}
}

/*


func (dbdata *xData) addSuperCourses() {
	dbp.SuperCourses = []db.SuperCourse{}
	dbdata.supercourses = map[Ref]db.DbRef{}
	for _, d := range dbdata.w365.SuperCourses {
		cr := dbdata.nextId()
		sr, ok := dbdata.subjects[d.Subject]
		if !ok {
			log.Printf("*ERROR* Unknown Subject in SuperCourse %s:\n  %s\n",
				d.Id, d.Subject)
			continue
		}
		dbp.SuperCourses = append(dbp.SuperCourses, db.SuperCourse{
			Id:        cr,
			Subject:   sr,
			Reference: string(d.Id),
		})
		dbdata.supercourses[d.Id] = cr
	}
}

func (dbdata *xData) addSubCourses() {
	dbp.SubCourses = []db.SubCourse{}
	dbdata.subcourses = map[Ref]db.DbRef{}
	for _, d := range dbdata.w365.SubCourses {
		sr, glist, tlist, rm := dbdata.readCourse(
			d.Id, d.Subject, d.Subjects, d.Groups, d.Teachers, d.PreferredRooms)
		sc, ok := dbdata.supercourses[d.SuperCourse]
		if !ok {
			log.Printf("*ERROR* Unknown SuperCourse in SubCourse %s:\n  %s\n",
				d.Id, d.SuperCourse)
			continue
		}
		cr := dbdata.nextId()
		dbp.SubCourses = append(dbp.SubCourses, db.SubCourse{
			Id:          cr,
			SuperCourse: sc,
			Subject:     sr,
			Groups:      glist,
			Teachers:    tlist,
			Room:        rm,
			Reference:   string(d.Id),
		})
		dbdata.subcourses[d.Id] = cr
	}
}

func (dbdata *xData) addLessons() {
	dbp.Lessons = []db.Lesson{}
	for _, d := range dbdata.w365.Lessons {
		// The course can be either a Course or a SubCourse.
		crs, ok := dbdata.courses[d.Course]
		if !ok {
			crs, ok = dbdata.subcourses[d.Course]
			if !ok {
				log.Printf("*ERROR* Invalid course in Lesson %s:\n  -- %s\n",
					d.Id, d.Course)
				continue
			}
		}
		rlist := []db.DbRef{}
		for _, r := range d.LocalRooms {
			rr, ok := dbdata.rooms[r]
			if ok {
				rlist = append(rlist, rr)
			} else {
				log.Printf("*ERROR* Invalid room in Lesson %s:\n  -- %s\n",
					d.Id, r)
			}
		}
		dbp.Lessons = append(dbp.Lessons, db.Lesson{
			Id:        dbdata.nextId(),
			Course:    crs,
			Duration:  d.Duration,
			Day:       d.Day,
			Hour:      d.Hour,
			Fixed:     d.Fixed,
			Rooms:     rlist,
			Reference: string(d.Id),
		})
	}
}
*/
