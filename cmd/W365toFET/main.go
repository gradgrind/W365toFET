/*
W365toFET produces a FET configuration file from a supplied Waldorf 365 data
set (JSON).

The name of the input file should ideally end with "_w365.json", for example
"myfile_w365.json". This will enable a consistent automatic naming of the
generated files.

The files produced are saved in the same directory as the input file:

  - Log file: Contains error messages and warnings as well as information
    about the steps performed. The file name (given "myfile_w365.json" as
    input) is "myfile_w365.log" – just the ending is changed.

  - FET file: The file to be fed to FET. The standard name (given
    "myfile_w365.json" as input) is "myfile.fet". However, by supplying a
    "FetFile" property (without the ".fet" ending) in the "FetData" object,
    this can be changed.

  - Map file: Correlates the FET Activity numbers to the Waldorf 365 Lesson
    references ("Id"). The standard name (given "myfile_w365.json" as input)
    is "myfile.map". However, by supplying a "MapFile" property
    (without the ".map" ending) in the "FetData" object, this can be changed.

Note that, at present, the Activity and Room objects in the FET file have the
corresponding Waldorf 365 references ("Id") in their "Comments" fields.

Firstly, the input file is read and processed so that the data can be stored
in a form independent of Waldorf 365. This form is managed in the [base]
package, the primary data structure being [base.DbTopLevel].

There are some useful pieces of information which are not stored directly
in the basic data loaded from an input file, but which can be derived from it.
The method [base.PrepareDb] performs the first of this processing and also
checks for certain errors in the data.

For processing of timetable information there are further useful data
structures which can be derived from the input data. This information is
handled primarily in the [ttbase] package, its primary data structure being
[ttbase.TtInfo].

A further stage of processing the timetable data is handled by the method
[ttbase.PrepareCoreData]. This builds further data structures representing
the allocation of resources, so that a number of errors in the data can be
detected.

Finally, the data structures are used by the function [fet.MakeFetFile] to
produce the XML-structure of the FET file and the reference mapping
information to be stored in the map file.
*/
package main

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/ttbase"
	"W365toFET/w365tt"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		if len(args) == 0 {
			log.Fatalln("ERROR* No input file")
		}
		log.Fatalf("*ERROR* Too many command-line arguments:\n  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("*ERROR* Couldn't resolve file path: %s\n", args[0])
	}

	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
	logpath := stempath + ".log"
	base.OpenLog(logpath)
	stempath = strings.TrimSuffix(stempath, "_w365")

	db := base.NewDb()
	w365tt.LoadJSON(db, abspath)
	db.PrepareDb()
	ttinfo := ttbase.MakeTtInfo(db)
	ttinfo.PrepareCoreData()

	// ********** Build the fet file **********

	fetfile := stempath
	mapfile := stempath
	thisdir := filepath.Dir(stempath)
	moduleData := ttinfo.Db.ModuleData
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

	xmlitem, lessonIdMap := fet.MakeFetFile(ttinfo)

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
}
