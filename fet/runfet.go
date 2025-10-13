package fet

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func Setup() {
	timetable.BACKEND = timetable.TtBackend{
		Run:     runFet,
		Abort:   ttRunAbort,
		Tick:    ttTick,
		Clear:   ttRunClear,
		Tidy:    ttRunTidy,
		Results: ttResults,
	}
}

func ttRunAbort(tt_data *timetable.TtData) {
	tt_data.BackEndData.(*fetTtData).cancel()
}

func ttRunTidy(workingdir string) {
	os.RemoveAll(filepath.Join(workingdir, "tmp"))
}

func ttRunClear(tt_data *timetable.TtData) {
	fttd, ok := tt_data.BackEndData.(*fetTtData)
	if ok {
		//base.Message.Printf("### Remove %s\n", fttd.workingdir)
		os.RemoveAll(fttd.workingdir)
		//} else {
		//	base.Message.Printf("### No TtData: %s\n", tt_data.Description)
	}
}

func runFet(tt_data *timetable.TtData, testing bool) {
	shared_data := tt_data.SharedData
	fname := tt_data.Description
	dir_n := filepath.Join(shared_data.WorkingDir, "tmp", fname)

	err := os.MkdirAll(dir_n, 0755)
	if err != nil {
		panic(err)
	}
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"

	// Construct the FET-file
	xmlitem := MakeFetFile(tt_data)

	// Write FET file
	err = os.WriteFile(fetfile, xmlitem, 0644)
	if err != nil {
		panic("Couldn't write fet output to: " + fetfile)
	}

	if tt_data.Description == "COMPLETE" {
		// Save fet file at top level of working directory.
		cfile := filepath.Join(shared_data.WorkingDir,
			filepath.Base(strings.TrimSuffix(
				shared_data.WorkingDir, "_fet")+".fet"))
		err = os.WriteFile(cfile, xmlitem, 0644)
		if err != nil {
			panic("Couldn't write fet output to: " + cfile)
		}
	}

	//fmt.Printf("FET file written to: %s\n", fetfile)

	/* TODO-- Convert lessonIdMap to string, write Id-map file.
	idmlines := []string{}
	for _, idm := range lessonIdMap {
		idmlines = append(idmlines,
			strconv.Itoa(int(idm.activityId))+":"+string(idm.baseId))
	}
	lidmap := strings.Join(idmlines, "\n")
	*/

	//TODO--
	//return

	cwd := filepath.Dir(fetfile)
	odir := filepath.Join(cwd, "out")
	os.RemoveAll(odir)
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	room_indexes := map[string]timetable.RoomIndex{}
	for i, rnode := range tt_data.SharedData.Db.Rooms {
		room_indexes[rnode.GetTag()] = timetable.RoomIndex(i)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &fetTtData{
		finished:   false,
		room2index: room_indexes,
		activities: len(shared_data.Activities),
		ifile:      fetfile,
		workingdir: cwd,
		odir:       odir,
		logfile:    logfile,
		cancel:     cancel,
	}
	tt_data.BackEndData = fet_data

	params := []string{
		"--inputfile=" + fetfile,
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
		"--outputdir=" + odir,
	}

	if testing {
		params = append(params,
			"--randomseeds10=10",
			"--randomseeds11=11",
			"--randomseeds12=12",
			"--randomseeds20=20",
			"--randomseeds21=21",
			"--randomseeds22=22")
	}

	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		"fet-cl", params...,
	)

	go run(fet_data, runCmd)
}

// The executable, `fet-cl`, places any messages in the `log` directory, as
// `result.txt` (which is probably not so interesting), `warnings.txt` (which
// might possibly containt something of diagnostic interest) and `errors.txt`
// (which may well contain diagnostic information that should ideally
// have been caught earlier ...). The `warnings.txt` and `errors.txt`
// files are present only if there is something to report. If there is an
// `errors.txt`, it should certainly be reported (in `fet_data.message`).

// The completion code of `fet-cl` is not particularly helpful, so the
// success of the run is determined by checking the number of placed
// activities and the existence of an `errors.txt` file.

// `run` is a goroutine. The last item to be changed must be `fet_data.state`,
// to avoid potential race conditions.
func run(fet_data *fetTtData, cmd *exec.Cmd) {
	cmd.CombinedOutput()
	fet_data.finished = true
}

var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

type fetTtData struct {
	activities int // total number of activities to place
	room2index map[string]timetable.RoomIndex
	ifile      string
	workingdir string
	odir       string
	logfile    string
	rdfile     *os.File // this must be closed when the subprocess finishes
	reader     *bufio.Reader
	cancel     func()
	finished   bool
}

// `ttTick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func ttTick(tt_data *timetable.TtData) {
	data := *tt_data.BackEndData.(*fetTtData)
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

							base.Report(fmt.Sprintf("%s: %d @ %d\n",
								tt_data.Description, percent, tt_data.Ticks))
						}
					}
				}
				break
			}
			panic(err)
		}
	}
exit:
	if data.finished {
		if data.rdfile != nil {
			data.rdfile.Close()
		}
		if tt_data.Progress == 100 {
			tt_data.State = 1
		} else {
			tt_data.State = 2
		}
		efile, err := os.ReadFile(filepath.Join(data.odir, "logs", "errors.txt"))
		if err != nil {
			tt_data.Message = string(efile)
		}
	}
}

func ttResults(tt_data *timetable.TtData) []timetable.ActivityPlacement {
	data := *tt_data.BackEndData.(*fetTtData)
	xmlpath := filepath.Join(data.odir, "timetables", tt_data.Description,
		tt_data.Description+"_activities.xml")
	// Open the XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		base.Bug.Print(err)
		return nil
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	base.Message.Printf("Reading: %s\n", xmlpath)
	byteValue, _ := io.ReadAll(xmlFile)
	v := fetResultRoot{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		base.Bug.Printf("XML error in %s:\n %v\n", xmlpath, err)
		return nil
	}

	activities := make([]timetable.ActivityPlacement, len(v.Activities))
	for i, a := range v.Activities {
		rooms := []timetable.RoomIndex{}
		if len(a.Real_Room) != 0 {
			for _, r := range a.Real_Room {
				rooms = append(rooms, data.room2index[r])
			}
		} else if len(a.Room) != 0 {
			rooms = append(rooms, data.room2index[a.Room])
		}
		activities[i] = timetable.ActivityPlacement{
			Id:    a.Id,
			Day:   a.Day,
			Hour:  a.Hour,
			Rooms: rooms,
		}
	}
	return activities
}

type fetResultRoot struct { // The root node.
	XMLName    xml.Name            `xml:"Activities_Timetable"`
	Activities []fetResultActivity `xml:"Activity"`
}

type fetResultActivity struct {
	XMLName   xml.Name `xml:"Activity"`
	Id        timetable.ActivityIndex
	Day       int
	Hour      int
	Room      string
	Real_Room []string
}
