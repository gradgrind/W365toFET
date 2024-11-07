package main

import (
	"W365toFET/fet"
	"W365toFET/w365tt"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	Message *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Bug     *log.Logger
)

func openlog(logpath string) {
	file, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	Message = log.New(file, "*INFO* ", log.Lshortfile)
	Warning = log.New(file, "*WARNING* ", log.Lshortfile)
	Error = log.New(file, "*ERROR* ", log.Lshortfile)
	Bug = log.New(file, "*BUG* ", log.Lshortfile)
}

func main() {

	//wordPtr := flag.String("word", "foo", "a string")
	//numbPtr := flag.Int("numb", 42, "an int")
	//forkPtr := flag.Bool("fork", false, "a bool")
	//var svar string
	//flag.StringVar(&svar, "svar", "bar", "a string var")

	flag.Parse()

	//fmt.Println("word:", *wordPtr)
	//fmt.Println("numb:", *numbPtr)
	//fmt.Println("fork:", *forkPtr)
	//fmt.Println("svar:", svar)
	fmt.Println("args:", flag.Args())

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("*ERROR* Too many command-line arguments:\n  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("*ERROR* Couldn't resolve file path: %s\n", args[0])
	}

	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
	logpath := stempath + ".log"
	openlog(logpath)

	data := w365tt.LoadJSON(abspath)

	// ********** Build the fet file **********
	stempath = strings.TrimSuffix(stempath, "_w365")

	xmlitem, lessonIdMap := fet.MakeFetFile(data)

	// Write FET file
	fetfile := stempath + ".fet"
	f, err := os.Create(fetfile)
	if err != nil {
		Bug.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		Bug.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	Message.Printf("FET file written to: %s\n", fetfile)

	//TODO: Write Id-map file.
	mapfile := stempath + ".map"
	fm, err := os.Create(mapfile)
	if err != nil {
		Bug.Fatalf("Couldn't open output file: %s\n", mapfile)
	}
	defer fm.Close()
	_, err = fm.WriteString(lessonIdMap)
	if err != nil {
		Bug.Fatalf("Couldn't write fet output to: %s\n", mapfile)
	}
	Message.Printf("Id-map written to: %s\n", mapfile)

	Message.Println("OK")
}
