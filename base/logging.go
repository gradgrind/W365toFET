package base

import (
	"fmt"
	"log"
	"os"
)

var (
	CONSOLE bool
	Message *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Bug     *log.Logger
)

func Report(msg string) {
	if CONSOLE {
		fmt.Print(msg)
	}
}

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

	Message = log.New(file, "*INFO* ", 0)
	Warning = log.New(file, "*WARNING* ", 0)
	Error = log.New(file, "*ERROR* ", 0)
	Bug = log.New(file, "*BUG* ", log.Lshortfile)

}

/* TODO: New error reporter?
func ERROR(msg string, args ...any) {
	fmt.Println("+++ Error +++++++++++++++")
	fmt.Printf(ErrorMessages[msg], args...)
	fmt.Println("\n-------------------------")
}
*/
