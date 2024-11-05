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
	data := w365tt.LoadJSON(abspath)

	// ********** Build the fet file **********
	//xmlitem := make_fet_file(&data, alist, course2activities, sgalist)
	xmlitem := fet.MakeFetFile(data)
	//fmt.Printf("\n*** fet:\n%v\n", xmlitem)
	fetfile := strings.TrimSuffix(abspath, filepath.Ext(abspath)) + ".fet"
	f, err := os.Create(fetfile)
	if err != nil {
		log.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		log.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	log.Printf("\nFET file written to: %s\n", fetfile)
}
