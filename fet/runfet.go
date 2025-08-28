package fet

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

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

	fmt.Printf("$ IN: %s\n", ifile)
	fmt.Printf("$ OUT: %s\n", odir)
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

	//TODO: Is this the right place to close the channel?
	// What if this goroutine is killed?
	//close(ch)
}

// Rather like a "tail" function, this can read the FET progress
// from its log file. It simply polls for new lines.

var pattern = "time (.*), FET reached ([0-9]+)"

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
