package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

type fetRun struct {
    output string
    err error
}

func main() {
    ifile := os.Args[1]
    odir := "output1"
    os.RemoveAll(odir)

    ch := make(chan fetRun)
    go runfet(ch, ifile, odir)

    logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")
    stop := make(chan bool)
    defer close(stop)
    finished := make(chan bool)
    go follow(stop, finished, logfile)

    fetresult := <-ch
    stop <- true
    _ = <-finished
    //close(stop)

    if err := fetresult.err; err != nil {
        switch e := err.(type) {
        case *exec.Error:
            fmt.Println("!!! failed executing:", err)
        case *exec.ExitError:
            fmt.Println("!!! command exit rc =", e.ExitCode())
        default:
            panic(err)
        }
    }
    fmt.Println("----->>>")
    fmt.Println(string(fetresult.output))
}

func runfet(ch chan fetRun, ifile string, odir string) {
    defer close(ch)
    runCmd := exec.Command(
        "fet-cl", "--inputfile=" + ifile,
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
    )

    res, err := runCmd.Output()
    ch <- fetRun{string(res), err}

    //TODO: Is this the right place to close the channel?
    // What if this goroutine is killed?
    //close(ch)
}
