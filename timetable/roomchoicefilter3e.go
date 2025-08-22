package main

import (
    "fmt"
    "slices"
    //"maps"
    //"strings"
    "runtime"
    "time"
)

type Ref = int

func roomChoiceFilter() {
    rlist := []Ref{
        //1,
    }

    rclist := [][]Ref{
        []Ref{1, 2},
        []Ref{3, 4, 5, 6},
        []Ref{7, 8, 9},
        []Ref{10, 11, 12},
        []Ref{13, 14, 12},
        []Ref{7, 8, 15},
        []Ref{9, 15},
        []Ref{16, 17, 9},
        []Ref{18, 19, 9},
        []Ref{18, 19, 15},
        []Ref{1, 2, 9},
        []Ref{16, 17, 12},
        []Ref{18, 19, 12},
        []Ref{20, 21},
        []Ref{7, 8, 12},
        []Ref{20, 21, 12},
        //[]Ref{16, 17, 15},
        []Ref{22, 23, 12},
        []Ref{3, 4, 9},
        []Ref{1, 2, 12},
    }

    type nothing = struct{}
    delta := 0

    necessary := slices.Clone(rlist)


//TODO: The rooms in a RoomChoiceGroup can perhaps be checked for
// validity when storing them in the db? Ordering can be done only
// when they have resource indexes – so that might have to stay here.
// If that is so, the checks can also stay here, I suppose (but as
// bug detectors?).

stage1:
    newlist := [][]Ref{}
    for i, rc0 := range rclist {
        // Filter out fixed rooms from the choice list
        rc := []Ref{}
        for _, r := range rc0 {
            if !slices.Contains(necessary, r) {
                rc = append(rc, r)
            }
        }
        if len(rc) >= 2 {
            // Sort the elements.
            slices.Sort(rc)
            fmt.Printf("$%d %v -> %v\n", i, rc0, rc)
            newlist = append(newlist, rc)
            fmt.Printf("(STATE1): [%d, %d, %d] %v\n",
                len(necessary), len(rclist), delta,
                necessary)
        } else if len(rc) == 1 {
            necessary = append(necessary, rc[0])
            fmt.Printf("$%d %v -> %d\n", i, rc0, rc[0])
            fmt.Printf("!!! FIXED %d, REPEATING\n", rc[0])
            rclist = append(newlist, rclist[i+1:]...)
            fmt.Printf("(STATE2): [%d, %d, %d] %v\n",
                len(necessary), len(rclist), delta,
                necessary)
            goto stage1
        } else {
            fmt.Printf("$%d %v -> {}\n", i, rc0)
            fmt.Printf("(STATE3): [%d, %d, %d] %v\n",
                len(necessary), len(rclist), delta,
                necessary)
            delta--
            if delta < 0  {
                fmt.Printf("ERROR: choice list %v has no new rooms\n", rc0)
                return
            }
        }
    }

    fmt.Printf("*******>>> %d %d %d\n", len(necessary), len(newlist), delta)

    // Now build the Cartesian product of the choice lists, omitting
    // values with duplicate rooms and duplicate values generally.
    cp := [][]Ref{{}} // build Cartesian product values here
    for i, rc := range newlist {
        // Add next choice list, extending the entries in `cp`
        newcp := [][]Ref{} // build new `cp` here
        for _, cp0 := range cp { // for each C-p value
            for _, r := range rc { // add each room in current choice list
                if !slices.Contains(cp0, r) { // ... if not a duplicate
                    cp1 := append(slices.Clone(cp0), r)
                    //fmt.Printf("???4: %v\n", cp1)

                    // ... and if the new C-p value is not a duplicate
                    slices.Sort(cp1)
                    //fmt.Printf("???5: %v\n", cp1)
                    for _, cp2 := range newcp {
                        if slices.Equal(cp2, cp1) {
                            goto next
                        }
                    }
                    newcp = append(newcp, cp1)
                    //fmt.Printf("???6: %v\n", newcp)
                next:
                }
            }
        }

        if len(newcp) == 0 {
            fmt.Printf("!!!!! newcp empty: %v\n", rc)
            return
        } else {
            do_restart := false
            for _, r := range rc {
                for _, cp0 := range newcp {
                    if !slices.Contains(cp0, r) {
                        goto next2
                    }
                }
                // r is in all combinations
                necessary = append(necessary, r)
                delta++
                fmt.Printf("!!! FIXED %d\n", r)
                do_restart = true
            next2:
            }
            if do_restart {
                fmt.Println(" ... restarting")
                rclist = newlist
                goto stage1
            }
        }

        cp = newcp
        fmt.Printf("???7: %d – %v\n", i, len(newcp))
        newcpx := newcp
        if len(newcpx) > 8 {
            newcpx = newcpx[:8]
        }
        //fmt.Printf("???7: %d – %d %v\n", i, len(newcp), newcpx)
        //PrintMemUsage()
    }

/*
    for _, cp0 := range cp {
        fmt.Printf(" === Cprod: %v\n", cp0)
    }
*/

    slices.Sort(necessary)
    fmt.Printf("\n $$ NECESSARY: %v\n\n", necessary)
    for i, rc := range newlist {
        rl := []Ref{}
        for _, r := range rc {
            rl = append(rl, r)
        }
        fmt.Printf("*** %d: %v\n", i, rl)
    }
    fmt.Printf("\n delta: %d\n", delta)

/*

restart:

    newlist := [][]Ref{}

    cp := []map[Ref]nothing{map[Ref]nothing{}}
    for i, rc0 := range rclist {
        rc := []Ref{}
        for _, r := range rc0 {
            if _, ok := necessary[r]; !ok {
                rc = append(rc, r)
            }
        }
        if len(rc) < 2 {
            if len(rc) == 1 {
                necessary[rc[0]] = nothing{}
                rclist = append(newlist, rclist[i+1:]...)
                fmt.Printf("Fixing %d\n", rc0[0])
                delta++
                goto restart // need to start again ...
            } else {
                fmt.Printf("Dropping %v\n", rc0)
                delta--
            }
            continue
        }

        newlist = append(newlist, rc)

        newcp := []map[Ref]nothing{}
        for _, cpm := range cp {
            for _, r := range rc {
                if _, ok := cpm[r]; !ok {
                    newcpm := map[Ref]nothing{r: nothing{}}
                    maps.Copy(newcpm, cpm)
                    newcp = append(newcp, newcpm)
                }
            }
        }
        cp = newcp

        // Look for "necessary" rooms
        for _, r := range rc {
            for _, cpm := range cp {
                if _, ok := cpm[r]; !ok {
                    goto next
                }
            }

            // The room is in all sets
            necessary[r] = nothing{}
            goto restart // need to start again ...

        next:
        }
    }
    rooms := []Ref{}
    for r := range necessary {
        rooms = append(rooms, r)
    }
    fmt.Printf("\n $$ NECESSARY: %s\n\n", strings.Join(rooms, ","))
    for i, rc := range newlist {
        rl := []Ref{}
        for _, r := range rc {
            rl = append(rl, r)
        }
        fmt.Printf("*** %d: %s\n", i, strings.Join(rl, ","))
    }

*/
}

func main() {
    start := time.Now()

    roomChoiceFilter()

    elapsed := time.Now().Sub(start)
    fmt.Printf("Run time: %s\n", elapsed)
}

func PrintMemUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("Alloc = %v MiB", m.Alloc / 1024 / 1024)
    //fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc / 1024 / 1024)
    fmt.Printf("\tSys = %v MiB", m.Sys / 1024 / 1024)
    //fmt.Printf("\tNumGC = %v", m.NumGC)
    fmt.Println()
}
