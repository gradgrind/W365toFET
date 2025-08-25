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
func SteerGeneration(tt_data *timetable.TtData, stempath string) {
	// Create a fresh working directory
	workingdir := stempath + "_fet"
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	run_number := 0
	// Each run of FET is in its own subdirectory

	run_number++
	dir_n := filepath.Join(workingdir, "run_"+strconv.Itoa(run_number))
	err = os.Mkdir(dir_n, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, filepath.Base(stempath))
	fetfile := stemfile + ".fet"
	mapfile := stemfile + ".map"
	xmlitem, lessonIdMap := fet.MakeFetFile(tt_data)

	// Write FET file
	f, err := os.Create(fetfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	base.Message.Printf("FET file written to: %s\n", fetfile)

	// Write Id-map file.
	fm, err := os.Create(mapfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", mapfile)
	}
	defer fm.Close()
	_, err = fm.WriteString(lessonIdMap)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", mapfile)
	}
	base.Message.Printf("Id-map written to: %s\n", mapfile)

	base.Message.Println("OK")

	// First run with no constraints except the hard-blocked time slots and
	// the fixed activities.

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

/*
// TODO: This was the original code to build the .fet and .map files.
// Is the bit fetching file names from db.ModuleData still needed somehow?

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

	xmlitem, lessonIdMap := fet.MakeFetFile(tt_data)

	// Write FET file
	f, err := os.Create(fetfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	base.Message.Printf("FET file written to: %s\n", fetfile)

	// Write Id-map file.
	fm, err := os.Create(mapfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", mapfile)
	}
	defer fm.Close()
	_, err = fm.WriteString(lessonIdMap)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", mapfile)
	}
	base.Message.Printf("Id-map written to: %s\n", mapfile)

	base.Message.Println("OK")

*/
