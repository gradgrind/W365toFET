package base

import (
	"log"
	"os"
)

var (
	Message *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Bug     *log.Logger
)

func OpenLog(logpath string) {
	var file *os.File
	if logpath == "" {
		file = os.Stderr
	} else {
		os.Remove(logpath)
		var err error
		file, err = os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	Message = log.New(file, "*INFO* ", log.Lshortfile)
	Warning = log.New(file, "*WARNING* ", log.Lshortfile)
	Error = log.New(file, "*ERROR* ", log.Lshortfile)
	Bug = log.New(file, "*BUG* ", log.Lshortfile)
}
