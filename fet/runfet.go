package fet

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func ttRunAbort(data any) {
	data.(fetTtData).cancel()
}

func NewFet(
	rundata *timetable.TtRunData,
	instance *timetable.TtInstance,
) {

	fname := "run_" + strconv.Itoa(rundata.RunCounter)
	dir_n := filepath.Join(rundata.WorkingDir, fname)
	err := os.Mkdir(dir_n, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"
	mapfile := stemfile + ".map"

	// Construct the FET-file
	xmlitem, lessonIdMap := MakeFetFile(rundata.TtData)

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

	cwd := filepath.Dir(fetfile)
	odir := filepath.Join(cwd, "out")
	os.RemoveAll(odir)
	run_name := filepath.Base(cwd)
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	instance.Abort = ttRunAbort
	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &fetTtData{
		logfile:  logfile,
		run_name: run_name,
		cancel:   cancel,
	}
	instance.HandlerData = fet_data

	//TODO: ctx needs to be passed to the exec command
	//??

	instance.UpdateHandler = ttUpdate

	//??

	ch := make(chan fetRun)

	//??

	go execfet(ch, fetfile, odir)

	//??

	stop := make(chan bool)
	defer close(stop)
	finished := make(chan bool)

	//??

	go follow(stop, finished, logfile, run_name)

	//??

	fetresult := <-ch
	stop <- true
	<-finished
	//close(stop)

	if err := fetresult.err; err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Printf("::%s>>> !!! failed executing: %s\n", run_name, err)
		case *exec.ExitError:
			// If FET aborts because of a data error, this case will be run
			// Is the exit code then always 1?
			// If killed the exit code seems to be -1.
			fmt.Printf("::%s>>> !!! command exit rc = %d\n",
				run_name, e.ExitCode())
		default:
			panic(err)
		}
	}
	fmt.Println("--------------------------------------------------")
	//fmt.Println("----->>>")
	//fmt.Println(string(fetresult.output))

}

//########################

type fetRun struct {
	output string
	err    error
}

func RunFet(fetfile string) {
	cwd := filepath.Dir(fetfile)
	odir := filepath.Join(cwd, "out")
	os.RemoveAll(odir)
	run_name := filepath.Base(cwd)

	ch := make(chan fetRun)
	go execfet(ch, fetfile, odir)

	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")
	stop := make(chan bool)
	defer close(stop)
	finished := make(chan bool)
	go follow(stop, finished, logfile, run_name)

	fetresult := <-ch
	stop <- true
	<-finished
	//close(stop)

	if err := fetresult.err; err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Printf("::%s>>> !!! failed executing: %s\n", run_name, err)
		case *exec.ExitError:
			// If FET aborts because of a data error, this case will be run
			// Is the exit code then always 1?
			// If killed the exit code seems to be -1.
			fmt.Printf("::%s>>> !!! command exit rc = %d\n",
				run_name, e.ExitCode())
		default:
			panic(err)
		}
	}
	fmt.Println("--------------------------------------------------")
	//fmt.Println("----->>>")
	//fmt.Println(string(fetresult.output))
}

func execfet(ch chan fetRun, ifile string, odir string) {
	defer close(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//fmt.Printf("$ IN: %s\n", ifile)
	//fmt.Printf("$ OUT: %s\n", odir)
	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		"fet-cl", "--inputfile="+ifile,
		"--writetimetablesstatistics=false",
		"--writetimetablesdayshorizontal=false",
		"--writetimetablesdaysvertical=false",
		"--writetimetablestimehorizontal=false",
		"--writetimetablestimevertical=false",
		"--writetimetablessubgroups=false",
		"--writetimetablesgroups=false",
		"--writetimetablesyears=false",
		"--writetimetablesteachers=false",
		"--writetimetablesteachersfreeperiods=false",
		"--writetimetablesbuildings=false",
		"--writetimetablesrooms=false",
		"--writetimetablessubjects=false",
		"--outputdir="+odir,
	)

	res, err := runCmd.Output()
	ch <- fetRun{string(res), err}
}

// Rather like a "tail" function, this can read the FET progress
// from its log file. It simply polls for new lines.

var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

type fetTtData struct {
	logfile  string
	rfile    *os.File // this must be closed when the subprocess finishes
	run_name string
	reader   *bufio.Reader
	cancel   func()
}

// TODO: need to tweak this because of the file close?
func initFetTtData(run_name string, ofile string) *fetTtData {
	file, err := os.Open(ofile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	return &fetTtData{
		run_name: run_name,
		reader:   bufio.NewReader(file),
	}
}

func ttUpdate(instance *timetable.TtInstance) {
	data := instance.HandlerData.(fetTtData)
	if data.reader == nil {
		// Await the existence of the log file
		file, err := os.Open(data.logfile)
		if err != nil {
			return
		}
		data.rfile = file // this needs closing
		data.reader = bufio.NewReader(file)
	}
	var l [][]byte
	for {
		line, err := data.reader.ReadString('\n')
		if err == nil {
			l = re.FindSubmatch([]byte(line))
			continue
		}
		if err == io.EOF {
			if l != nil {
				//TODO
				fmt.Printf(" .. %s> %s : %s\n",
					data.run_name, string(l[1]), string(l[2]))
			}
			//TODO: there may need to be a special action here when
			// the subprocess has finished.
			return
		}
		panic(err)
	}
}

func follow(
	stop chan bool, finished chan bool, ofile string, run_name string) {

	re := regexp.MustCompile(pattern)

	defer close(finished)

	var reader *bufio.Reader
	open := false
	done := false

	for {
		select {
		case <-stop:
			done = true
		default:
			if open {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						if done {
							finished <- true
							//close(finished)
							return
						}
						time.Sleep(200 * time.Millisecond)
						continue
					}
					fmt.Println(err)
					finished <- false
					//close(finished)
					return
				}
				l := re.FindSubmatch([]byte(line))
				if l != nil {
					fmt.Printf(" .. %s> %s : %s\n",
						run_name, string(l[1]), string(l[2]))
				}
			} else {
				file, err := os.Open(ofile)
				if err != nil {
					time.Sleep(100 * time.Millisecond)
				} else {
					defer file.Close()
					reader = bufio.NewReader(file)
					open = true
				}
			}
		}
	}
}
