package fet

import (
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
)

func ttRunAbort(data any) {
	data.(fetTtData).cancel()
}

func NewFet(instance *timetable.TtInstance) {
	fname := instance.Description
	dir_n := filepath.Join(instance.WorkingDir, fname)
	err := os.Mkdir(dir_n, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"
	mapfile := stemfile + ".map"

	// Construct the FET-file
	xmlitem, lessonIdMap := MakeFetFile(instance.TtData)

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
	//fmt.Printf("FET file written to: %s\n", fetfile)

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
	//fmt.Printf("Id-map written to: %s\n", mapfile)

	cwd := filepath.Dir(fetfile)
	odir := filepath.Join(cwd, "out")
	os.RemoveAll(odir)
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	instance.Abort = ttRunAbort
	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &fetTtData{
		activities: len(instance.TtData.Activities),
		ifile:      fetfile,
		odir:       odir,
		logfile:    logfile,
		cancel:     cancel,
	}
	instance.HandlerData = fet_data
	instance.UpdateHandler = ttUpdate
	go execfet(ctx, fet_data)
}

func execfet(
	ctx context.Context,
	fet_data *fetTtData,
) {
	//fmt.Printf("$ IN: %s\n", ifile)
	//fmt.Printf("$ OUT: %s\n", odir)
	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		"fet-cl", "--inputfile="+fet_data.ifile,
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
		"--outputdir="+fet_data.odir,
	)

	res, err := runCmd.Output()

	if err == nil {
		fet_data.state = timetable.NewState{
			State: 1, Message: string(res)}
	} else {
		switch e := err.(type) {
		case *exec.Error:
			panic(fmt.Sprintf(
				">>> !!! Failed running FET on %s:\n  %s\n",
				fet_data.ifile, err))
		case *exec.ExitError:
			// If FET aborts because of a data error, this case will be run
			// Is the exit code then always 1?
			// If killed the exit code seems to be -1.
			fmt.Printf(">>> !!! FET cc on %s = %d\n",
				fet_data.ifile, e.ExitCode())
			if e.ExitCode() < 0 {
				// aborted
				fet_data.state = timetable.NewState{
					State: 3, Message: string(res)}
			} else {
				// error completion
				fet_data.state = timetable.NewState{
					State: 2, Message: string(res)}
			}
		default:
			panic(err)
		}
	}
}

// Rather like a "tail" function, this can read the FET progress
// from its log file. It simply polls for new lines.

var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

type fetTtData struct {
	state      timetable.NewState
	activities int // total number of activities to place
	ifile      string
	odir       string
	logfile    string
	rdfile     *os.File // this must be closed when the subprocess finishes
	reader     *bufio.Reader
	cancel     func()
}

// `ttUpdate` runs in the event loop, so it may update the instance data.
// It is called on every "tick".
func ttUpdate(instance *timetable.TtInstance) {
	data := instance.HandlerData.(fetTtData)
	finished := data.state.State > 0
	if data.reader == nil {
		// Await the existence of the log file
		file, err := os.Open(data.logfile)
		if err != nil {
			goto exit
		}
		data.rdfile = file // this needs closing
		data.reader = bufio.NewReader(file)
	}
	{
		var l [][]byte
		for {
			line, err := data.reader.ReadString('\n')
			if err == nil {
				l = re.FindSubmatch([]byte(line))
				continue
			}
			if err == io.EOF {
				if l != nil {
					count, err := strconv.Atoi(string(l[2]))
					if err == nil {
						percent := count * 100 / len(instance.TtData.Activities)
						if percent > instance.Progress {
							instance.Progress = percent
							instance.LastTime = instance.Ticks
						}
					}
				}
				break
			}
			panic(err)
		}
	}
exit:
	if finished {
		if data.rdfile != nil {
			data.rdfile.Close()
		}
		instance.State = data.state.State
		instance.Message = data.state.Message
	}
}
