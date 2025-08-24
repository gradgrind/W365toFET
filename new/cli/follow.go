package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "time"
    "regexp"
)

// Rather like a "tail" function, this can read the FET progress
// from its log file. It simply polls for new lines.

var pattern = "time (.*), FET reached ([0-9]+)"

func follow(stop chan bool, finished chan bool, ofile string) {
    re := regexp.MustCompile(pattern)

    defer close(finished)

    var reader *bufio.Reader
    open := false
    done := false

    for {
        select {
        case <- stop:
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
                    fmt.Printf(" @ %s : %s\n", string(l[1]), string(l[2]))
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
