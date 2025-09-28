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

func Setup() {
	timetable.BACKEND = timetable.TtBackend{
		Run:   runFet,
		Abort: ttRunAbort,
		Tick:  ttTick,
	}
}

func ttRunAbort(tt_data *timetable.TtData) {
	tt_data.BackEndData.(*fetTtData).cancel()
}

func runFet(tt_data *timetable.TtData) {
	shared_data := tt_data.SharedData
	fname := tt_data.Description
	dir_n := filepath.Join(shared_data.WorkingDir, fname)

	//err := os.MkdirAll(newpath, os.ModePerm)
	err := os.Mkdir(dir_n, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"
	mapfile := stemfile + ".map"

	// Construct the FET-file
	xmlitem, lessonIdMap := MakeFetFile(tt_data)

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

	//TODO--
	//return

	cwd := filepath.Dir(fetfile)
	odir := filepath.Join(cwd, "out")
	os.RemoveAll(odir)
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &fetTtData{
		state:      0,
		activities: len(shared_data.Activities),
		ifile:      fetfile,
		odir:       odir,
		logfile:    logfile,
		cancel:     cancel,
	}
	tt_data.BackEndData = fet_data

	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		"fet-cl", "--inputfile="+fetfile,
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

	go run(fet_data, runCmd)
}

//TODO: fet-cl places any messages in the log directory, as result.txt
// (which is probably not so interesting), warnings.txt (which might
// possibly containt something of diagnostic interest) and errors.txt
// (which may well contain diagnostic information that should ideally
// have been caught earlier ...). The warnings.txt and errors.txt
// files may not be present. If there is an errors.txt, it should
// certainly be reported somehow (in fet_data.message?).

// The last item to be changed must be `fet_data.state`, to avoid slightly
// possible race conditions.
func run(fet_data *fetTtData, cmd *exec.Cmd) {
	_, err := cmd.CombinedOutput()
	if err == nil {
		// Apparently this can happen even if the generation is not
		// complete, so the handler (`ttTick`, see below) in the tick-loop
		// should check for completion.
		fet_data.state = 1
	} else {
		switch e := err.(type) {
		case *exec.Error:
			panic(fmt.Sprintf(
				">>> !!! Failed running FET on %s:\n  %s\n",
				fet_data.ifile, err))
		case *exec.ExitError:
			// If FET fails because of a data error, this case will be run,
			// apparently with exit code = 1 (not sure if this is always
			// the case).
			// If terminated by a timeout the exit code seems to be -1.
			fmt.Printf(">>> !!! FET cc on %s = %d\n",
				fet_data.ifile, e.ExitCode())
			if e.ExitCode() < 0 {
				// aborted
				fet_data.state = 3
			} else {
				// error completion
				fet_data.state = 2
			}
		default:
			panic(err)
		}
	}
}

var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

type fetTtData struct {
	state      int
	activities int // total number of activities to place
	ifile      string
	odir       string
	logfile    string
	rdfile     *os.File // this must be closed when the subprocess finishes
	reader     *bufio.Reader
	cancel     func()
}

// `ttTick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func ttTick(tt_data *timetable.TtData) {
	tt_data.Ticks++
	data := tt_data.BackEndData.(*fetTtData)
	finished := data.state > 0
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
						percent := count * 100 /
							(len(tt_data.SharedData.Activities) - 1)
						if percent > tt_data.Progress {
							tt_data.Progress = percent
							tt_data.LastTime = tt_data.Ticks

							//TODO
							fmt.Println(tt_data.Description, percent, "@", tt_data.Ticks)
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
		if data.state == 1 {
			// cc = 0 does not absolutely guarantee that the timetable is
			// complete – when fet-cl is interrupted, for example
			if tt_data.Progress == 100 {
				tt_data.State = 1
			} else {
				tt_data.State = 4
			}
		} else {
			tt_data.State = data.state
		}

		efile, err := os.ReadFile(filepath.Join(data.odir, "logs", "errors.txt"))
		if err != nil {
			tt_data.Message = string(efile)
		}
	}

	fmt.Printf("??? %s %d @ %d\n", tt_data.Description, tt_data.State, tt_data.Ticks)

}
