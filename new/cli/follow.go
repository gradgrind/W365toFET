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

func main() {
    re := regexp.MustCompile(pattern)

	file, err := os.Open("logs/max_placed_activities.txt")
	if err != nil {
		return
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			break
		}
		
		l := re.FindSubmatch([]byte(line))
		if l != nil {
			fmt.Printf(" @ %s : %s\n", string(l[1]), string(l[2]))
		}
	}
}
