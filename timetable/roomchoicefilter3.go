package main

import (
    "fmt"
    "slices"
    "maps"
    "strings"
    "runtime"
    "time"
)

type Ref = string

func roomChoiceFilter() {
    rlist := []Ref{
        //"ph1",
    }

    rclist := [][]Ref{
        []Ref{"ph1", "ph2"},
        []Ref{"mu1", "mu2", "mu3", "mu4"},
        []Ref{"m1", "m2", "k11a"},
        []Ref{"ml1", "ml2", "k7b"},
        []Ref{"geo1", "geo2", "k7b"},
        []Ref{"m1", "m2", "k11b"},
        []Ref{"k11a", "k11b"},
        []Ref{"b1", "b2", "k11a"},
        []Ref{"d1", "d2", "k11a"},
        []Ref{"d1", "d2", "k11b"},
        []Ref{"ph1", "ph2", "k11a"},
        []Ref{"b1", "b2", "k7b"},
        []Ref{"d1", "d2", "k7b"},
        []Ref{"ch1", "ch2"},
        []Ref{"m1", "m2", "k7b"},
        []Ref{"ch1", "ch2", "k7b"},
        //[]Ref{"b1", "b2", "k11b"},
        []Ref{"g1", "g2", "k7b"},
        []Ref{"mu1", "mu2", "k11a"},
        //[]Ref{"ph1", "ph2", "k7b"},
    }

    type nothing = struct{}
    delta := 0

    necessary := map[Ref]nothing{}
    for _, r := range rlist {
        necessary[r] = nothing{}
    }


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
            if _, ok := necessary[r]; !ok {
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
                slices.Sorted(maps.Keys(necessary)))
        } else if len(rc) == 1 {
            necessary[rc[0]] = nothing{}
            fmt.Printf("$%d %v -> %s\n", i, rc0, rc[0])
            fmt.Printf("!!! FIXED %s, REPEATING\n", rc[0])
            rclist = append(newlist, rclist[i+1:]...)
            fmt.Printf("(STATE2): [%d, %d, %d] %v\n",
                len(necessary), len(rclist), delta,
                slices.Sorted(maps.Keys(necessary)))
            goto stage1
        } else {
            fmt.Printf("$%d %v -> {}\n", i, rc0)
            fmt.Printf("(STATE3): [%d, %d, %d] %v\n",
                len(necessary), len(rclist), delta,
                slices.Sorted(maps.Keys(necessary)))
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
            for _, r := range newcp[0] {
                for _, cp0 := range newcp[1:] {
                    if !slices.Contains(cp0, r) {
                        goto next2
                    }
                }
                // r is in all combinations
                necessary[r] = nothing{}
                delta++
                fmt.Printf("!!! FIXED %s, restarting\n", r)
                rclist = newlist
                goto stage1
            next2:
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
                fmt.Printf("Fixing %s\n", rc0[0])
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
