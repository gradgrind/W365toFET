package base

import (
	"fmt"
	"log"
	"os"
)

// TODO: Gradually replace these loggers by ReportError (etc.)
var (
	Message *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Bug     *log.Logger

	logbase *LogBase
)

type LogBase struct {
	Logger   *log.Logger
	LangMap  map[string]string
	Fallback map[string]string
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

	logbase.Logger = log.New(file, "++", log.Lshortfile)

	Message = log.New(file, "*INFO* ", log.Lshortfile)
	Warning = log.New(file, "*WARNING* ", log.Lshortfile)
	Error = log.New(file, "*ERROR* ", log.Lshortfile)
	Bug = log.New(file, "*BUG* ", log.Lshortfile)
}

// TODO ... Expect a string array, to allow for some type checking, even
// though an any array is actually needed.
func (logbase *LogBase) ReportError(msg string, args ...string) {
	// Look up message
	msgt, ok := logbase.LangMap[msg]
	if !ok {
		msgt, ok = logbase.Fallback[msg]
		if !ok {
			logbase.Logger.Printf("Unknown message: %s ::: %+v\n",
				msg, args)
			panic("Unknown message")
		}
	}

	vlist := []any{}
	for _, arg := range args {
		vlist = append(vlist, arg)
	}
	logbase.Logger.Printf(msgt+"\n", vlist...)
}

// Tr adds message strings to the Fallback map of logbase, initializing
// logbase if necessary.
// It can be called from init functions.
func Tr(trmap map[string]string) {
	lg := logbase
	if lg == nil {
		lg = &LogBase{
			//Logger: log.New(file, "++", log.Lshortfile),
			//TODO: load the data from somewhere ...
			LangMap:  map[string]string{},
			Fallback: map[string]string{},
		}
		logbase = lg
	}
	for k, v := range trmap {
		if _, nok := lg.Fallback[k]; nok {
			fmt.Printf("Message defined twice: %s\n", k)
			panic("Message defined twice")
		}
		lg.Fallback[k] = v
	}
}
