package ttengine

import (
	"W365toFET/base"
	"W365toFET/readxml"
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

var inputfiles = []string{
	"../testdata/readxml/Demo1.xml",
	"../testdata/readxml/x01.xml",
}

func test1() {
	N := 1000000
	// (0), 1:0, 2:1, 3:2, 4:5, 5:7, 6:0, 7:4, 8:1
	list := []Penalty{-1, -1, -1, 0, 2, 7, 14, 14, 18, 19}
	rmap := map[int]int{}
	for c := 0; c < N; c++ {
		r := rand.IntN(20)
		i, _ := slices.BinarySearch(list, Penalty(r))
		//fmt.Printf(" ??? %d --> %d\n", r, i)
		rmap[i]++
	}
	fmt.Println(" -->")
	for i := 0; i < 10; i++ {
		n, ok := rmap[i]
		if ok {
			fmt.Printf(" %d: %+f\n", i-1, float32(n)*20.0/float32(N))
		}
	}
	fmt.Println()
}

func test2() {
	var total Penalty = -1
	penalties := []Penalty{0, 1, 7, 9, 3, 5, 0, 8, 4}
	pvec := make([]Penalty, len(penalties))
	for i, p := range penalties {
		total += p
		pvec[i] = total
	}
	fmt.Printf("::: %+v\n", pvec)
	N := 1000000
	rmap := map[int]int{}
	ILIM := int(total) + 1
	for c := 0; c < N; c++ {
		r := rand.IntN(ILIM)
		i, _ := slices.BinarySearch(pvec, Penalty(r))
		//fmt.Printf(" ??? %d --> %d\n", r, i)
		rmap[i]++
	}
	fmt.Println(" -->")
	for i := 0; i < 10; i++ {
		n, ok := rmap[i]
		if ok {
			fmt.Printf(" %d: %+f\n", i, float32(n)*float32(ILIM)/float32(N))
		}
	}
	fmt.Println()
}

func TestTtEngine(t *testing.T) {
	test2()
	return

	base.OpenLog("")
	for _, fxml := range inputfiles {
		fmt.Println("\n ++++++++++++++++++++++")
		cdata := readxml.ConvertToDb(fxml)
		fmt.Println("*** Available Schedules:")
		slist := cdata.ScheduleNames()
		for _, sname := range slist {
			fmt.Printf("  -- %s\n", sname)
		}
		sname := "Vorlage"
		if !slices.Contains(slist, sname) {
			if len(slist) != 0 {
				sname = slist[0]
			} else {
				fmt.Println(" ... stopping ...")
				continue
			}
		}
		fmt.Printf("*** Using Schedule '%s'\n", sname)
		if !cdata.ReadSchedule(sname) {
			fmt.Println(" ... failed ...")
			continue
		}

		db := cdata.Db()
		db.PrepareDb()
		ttbase.MakeTtInfo(db)
		ttinfo := ttbase.MakeTtInfo(db)
		ttinfo.PrepareCoreData()

		PlaceLessons(ttinfo)
	}
}
