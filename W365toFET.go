package main

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/ttbase"
	"W365toFET/ttprint"
	"W365toFET/w365tt"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	//wordPtr := flag.String("word", "foo", "a string")
	//numbPtr := flag.Int("numb", 42, "an int")
	//forkPtr := flag.Bool("fork", false, "a bool")
	//var svar string
	//flag.StringVar(&svar, "svar", "bar", "a string var")

	printflag := flag.Bool("p", false, "print timetables")

	flag.Parse()

	//fmt.Println("word:", *wordPtr)
	//fmt.Println("numb:", *numbPtr)
	//fmt.Println("fork:", *forkPtr)
	//fmt.Println("svar:", svar)
	//fmt.Println("args:", flag.Args())

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

	if *printflag {
		datadir := filepath.Join(filepath.Dir(abspath), "data")
		ttprint.PrintTimetables(ttinfo, datadir, stempath)
		base.Message.Println("OK")
		return
	}

	// ********** Build the fet file **********

	xmlitem, lessonIdMap := fet.MakeFetFile(ttinfo)

	// Write FET file
	fetfile := stempath + ".fet"
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
	mapfile := stempath + ".map"
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
