package autotimetable

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/timetable"
	"os"
	"path/filepath"
	"strconv"
)

// TODO: At present this only supports a FET back-end. Perhaps a choice should
// be possible ...
//
/*
// TODO:  Is this old bit fetching file names from db.ModuleData still
// needed somehow?

	fetfile := stempath
	mapfile := stempath
	thisdir := filepath.Dir(stempath)
	moduleData := db.ModuleData
	fetData, ok := moduleData["FetData"].(map[string]string)
	if ok {
		var f string
		f, ok = fetData["FetFile"]
		if ok {
			fetfile = filepath.Join(thisdir, f)
		}
		f, ok = fetData["MapFile"]
		if ok {
			mapfile = filepath.Join(thisdir, f)
		}
	}
	fetfile += ".fet"
	mapfile += ".map"
*/

var run_number int

func SteerGeneration(tt_data_0 *timetable.TtData, stempath string) {
	// `stempath` provides the path to the source file, including the stem
	// (without file-type extension) of the file name. A new working directory
	// will be created in the same directory.
	workingdir := stempath + "_fet"
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	////TODO-- This is just for testing
	//run_number = -1 // each run of FET is in its own subdirectory
	//runFET(tt_data_0, workingdir)
	////TODO++ This is the normal version
	run_number = 0 // each run of FET is in its own subdirectory

	// Copy original DbTopLevel (shallow copy only!)
	db0 := tt_data_0.Db
	db_1 := *db0
	db := &db_1

	// Copy original TtData (shallow copy only!)
	tt_data_1 := *tt_data_0
	tt_data := &tt_data_1
	tt_data.Db = db

	// Remove constraints
	tt_data.Constraints = map[string][]any{}
	tt_data.MinDaysBetweenLessons = nil
	tt_data.ParallelLessons = nil
	tt_data.WITHOUT_ROOM_PLACEMENTS = true

	// First run with no constraints except the hard-blocked time slots and
	// the fixed activities.
	// Regenerate the teachers list without constraints
	new_teachers := make([]*base.Teacher, len(db.Teachers))
	for i, t0p := range db.Teachers {
		t := *t0p
		t.MinLessonsPerDay = -1 // unconstrained
		t.MaxLessonsPerDay = -1 // unconstrained
		t.MaxDays = -1          // unconstrained
		t.MaxGapsPerDay = -1    // unconstrained
		t.MaxGapsPerWeek = -1   // unconstrained
		t.MaxAfternoons = -1    // unconstrained
		t.LunchBreak = false
		new_teachers[i] = &t
	}
	db.Teachers = new_teachers
	// Regenerate the classes list without constraints
	new_classes := make([]*base.Class, len(db.Classes))
	for i, c0p := range db.Classes {
		c := *c0p
		c.MinLessonsPerDay = -1 // unconstrained
		c.MaxLessonsPerDay = -1 // unconstrained
		c.MaxGapsPerDay = -1    // unconstrained
		c.MaxGapsPerWeek = -1   // unconstrained
		c.MaxAfternoons = -1    // unconstrained
		c.LunchBreak = false
		c.ForceFirstHour = false
		new_classes[i] = &c
	}
	db.Classes = new_classes

	//TODO... ???

	runFET(tt_data, workingdir)

	//// Restore original data
	//tt_data.Constraints = tt_data_0.Constraints
	//tt_data.MinDaysBetweenLessons = tt_data_0.MinDaysBetweenLessons
	//tt_data.ParallelLessons = tt_data_0.ParallelLessons
	//tt_data.WITHOUT_ROOM_PLACEMENTS = tt_data_0.WITHOUT_ROOM_PLACEMENTS
	//tt_data.Db = db0

	// A run with all constraints enabled
	runFET(tt_data_0, workingdir)

	/* ???

	// Add the pseudo activities due to the NotAvailable lists of classes,
	// teachers and rooms.
	tt_data.BlockResources()

	TODO?
	// Get preliminary constraint info – needed for the call to addActivity
	//ttinfo.processConstraints()

	TODO?
	// Add the remaining Activity information
	//ttinfo.addActivityInfo(t2tt, r2tt, g2ags)

	*/
}

func runFET(tt_data *timetable.TtData, workingdir string) {
	run_number++
	fname := "run_" + strconv.Itoa(run_number)
	dir_n := filepath.Join(workingdir, fname)
	err := os.Mkdir(dir_n, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"
	mapfile := stemfile + ".map"

	// Construct the FET-file
	xmlitem, lessonIdMap := fet.MakeFetFile(tt_data)

	// Write FET file
	f, err := os.Create(fetfile)
	if err != nil {
		panic("Couldn't open output file: " + fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		panic("Couldn't write fet output to: " + fetfile)
	}
	base.Message.Printf("FET file written to: %s\n", fetfile)

	// Write Id-map file.
	fm, err := os.Create(mapfile)
	if err != nil {
		panic("Couldn't open output file: " + mapfile)
	}
	defer fm.Close()
	_, err = fm.WriteString(lessonIdMap)
	if err != nil {
		panic("Couldn't write fet output to: " + mapfile)
	}
	base.Message.Printf("Id-map written to: %s\n", mapfile)

	base.Message.Println("OK")

	fet.RunFet(fetfile)
}
