package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"syscall"
	"time"
)

var (
	// The behaviour of the TESTING flag depends on the back-end. It might,
	// for example, use fixed seeds for random number generators so as to
	// produce reproduceable runs.
	TESTING bool
	// This approach relies on parallel processing. If there are too few real
	// processors it will be inefficient:
	MAXPROCESSES int

	NEW_BASE_TIMEOUT_FACTOR  int // factor * 10
	STAGE_TIMEOUT_MIN        int
	STAGE_TIMEOUT            int
	NEW_STAGE_TIMEOUT_FACTOR int // factor * 10

	DEBUG bool

	InstanceCounter         int = 0
	LastResult              *Result
	UNCHANGED_LIMIT_PERCENT int
)

func SetParameterDefault() {
	MAXPROCESSES = min(max(runtime.NumCPU(), 4), 6)

	NEW_BASE_TIMEOUT_FACTOR = 12 // => 1.2
	STAGE_TIMEOUT_MIN = 5
	//NEW_STAGE_TIMEOUT_FACTOR = 20 // => 2.0
	NEW_STAGE_TIMEOUT_FACTOR = 12 // => 1.5
	UNCHANGED_LIMIT_PERCENT = 80

	DEBUG = false
}

func init() {
	SetParameterDefault()
}

/*
Various strategies are used to try to achieve a – possibly imperfect –
timetable within a specified time. It is impossible to guarantee that all
constraints will be satisfied within a given time, so in order to place
all the activities within this time it may be necessary to drop some of
them.

A certain degree of parallel processing is assumed – too few (less than four?)
processor cores is likely to result in a very significant slowdown.

TODO: hard-only, with/without rooms, additional fully constrained? Perhaps
dependent on number of cores?

The main function (`StartGeneration`) starts a run with the fully constrained
data and a second run with all the "non-basic" constraints removed. Fixed
activity placements and blocked time-slots (for teachers, classes, and rooms)
are regarded as basic, non-negotiable.

A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine. Each instance
has its own individual timeout. There is also a global timeout to stop
all instances which are still running.

Once these initial instances have been started, a "tick-loop" (which is
triggered every second) is entered. This monitors the progress of each active
instance and handles the actions resulting from their completion, whether
successful or not.

Should a fully constrained instance complete successfully within the
allotted time, all other instances are terminated and its result will be
saved.

When the unconstrained instance completes successfully, a series of further
instances is queued for running, each specifying the addition of a list of
(hard) constraints of a single type. Thus for each type of constraint an
instance is constructed. Using timeouts leading to binary divisions of these
lists an attempt is made to find individual "difficult" constraints, which can
then be disabled in order to get full activity placement within a reasonable
time. Parallel processing can be of some assistance here.

TODO: Should the unconstrained instance fail to complete successfully within
its allotted time, further steps may be taken to trace difficulties within the
activity collection, perhaps identifying "difficult" classes or teachers.

When a single-constraint-type instance completes successfully, it is used as
a new base (`current_instance`) for the addition of further constraints. All
the remaining constraint-type instances are stopped and restarted with this
new base. If a constraint-type instance is timed out, it is stopped and split
into two halves, which then run in its place. If there are no halves (only
one constraint being added) there is no successor, the constraint is dropped.

When an instance completes successfully within the allotted time, its result
is saved as a JSON file, so that the best result so far gradually encompasses
more of the constraints. However, it can happen that the divisions complete
before the overall timeout occurs, leaving only the fully constrained
instance running.

TODO: At this point rejected constraints should be tried again, but with
longer timeouts.

The results include diagnostic information (at least an indication of which
constraints were dropped).

If the first run through of the single-constraint accumulation doesn't
complete before the overall timeout, that may indicate that a longer overall
timeout might be considered.

TODO: There is probably no general "optimum" value for the various timeouts,
that is likely to depend on the data. But perhaps values can be found which
are frequently useful. It might be helpful to use shorter overall timeouts
during the initial phases of testing the data, to identify potential problem
areas without long processing delays. For later phases longer times may be
necessary (depending on the difficulty of the data).
*/

var Descriptions map[string]string = map[string]string{
	"COMPLETE":           "All constraints active",
	"ONLY_BLOCKED_SLOTS": "All constraints – except blocked slots – disabled",
}

func StartGeneration(tt_data_0 *timetable.TtData, TIMEOUT int) {
	LastResult = nil
	tt_shared_data := tt_data_0.SharedData

	clashes := test_fixed(tt_data_0)
	if len(clashes) != 0 {
		for _, clash := range clashes {
			if clash.Course1 == nil {
				base.Error.Printf(
					"(TODO) Fixed lesson in blocked slot: %s @ %d.%d,\n Course %s\n",
					clash.Resource.GetResourceTag(),
					clash.Slot.Day,
					clash.Slot.Hour,
					tt_shared_data.View(clash.Course2),
				)

			} else {
				base.Error.Printf(
					"(TODO) Fixed lesson clash: %s @ %d.%d,\n Courses %s & %s\n",
					clash.Resource.GetResourceTag(),
					clash.Slot.Day,
					clash.Slot.Hour,
					tt_shared_data.View(clash.Course1),
					tt_shared_data.View(clash.Course2),
				)
			}
		}
		return
	}

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	runqueue := &RunQueue{
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: MAXPROCESSES,
		Next:       0,
	}

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist: existing
	// contents need not be preserved during processing.
	workingdir := tt_shared_data.WorkingDir

	tt_data_0.Description = "COMPLETE"

	// Global data
	Ticks = 0
	TtData_0 = tt_data_0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	full_instance := &TtInstance{
		Timeout: 0,
		TtData:  tt_data_0,
	}

	// Add to run queue
	runqueue.add(full_instance)

	//TODO: Instance without soft constraints (if any)?

	// Unconstrained instance
	STAGE_TIMEOUT = STAGE_TIMEOUT_MIN
	tt_data := new_ttdata(tt_data_0, "ONLY_BLOCKED_SLOTS")
	null_instance := &TtInstance{
		//Timeout: max(TIMEOUT/UNCONSTRAINED_TIMEOUT_FRACTION,
		//	MIN_UNCONSTRAINED_TIMEOUT),
		Timeout: STAGE_TIMEOUT,

		TtData: tt_data,
		HardConstraintEnabled: setup_hard_constraint_map(
			tt_data_0.HardConstraints),
	}
	disable_all_constraints(null_instance.TtData)
	// Add to run queue
	runqueue.add(null_instance)

	// Start stage 0
	stage := 0
	full_progress := 0      // current percentage
	full_progress_last := 0 // time of last increment

	// *** Ticker loop ***
	var constraint_list []*TtInstance
	var sidelined []*TtInstance
	var current_instance *TtInstance
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer func() {
		// Tidy up
		r := recover()
		if r != nil {
			base.Message.Println("(TODO) *** RECOVER ***", r)
			fmt.Printf("(TODO) *** RECOVER *** %s\n%s\n",
				r, debug.Stack())
		}
		for {
			count := 0
			for instance := range runqueue.Active {
				if instance.TtData.State == 0 {
					timetable.BACKEND.Tick(instance.TtData)
					count++
					abort_instance(instance)
				}
			}
			if count == 0 {
				break
			}
			<-ticker.C
		}
		if !DEBUG {
			// Remove all remaining temporary files
			timetable.BACKEND.Tidy(workingdir)
		}
		if LastResult != nil {
			//b, err := json.Marshal(LastResult)
			b, err := json.MarshalIndent(LastResult, "", "  ")
			if err != nil {
				panic(err)
			}
			fpath := filepath.Join(workingdir, "Result.json")
			f, err := os.Create(fpath)
			if err != nil {
				panic("Couldn't open output file: " + fpath)
			}
			defer f.Close()
			_, err = f.Write(b)
			if err != nil {
				panic("Couldn't write result to: " + fpath)
			}
		}
	}()

tickloop:
	for runqueue.update_queue() != 0 || stage != 3 {
		select {
		case <-ticker.C:
		case <-sigChan:
			base.Message.Printf("(TODO) *** INTERRUPTED @ %d ***\n", Ticks)
			break tickloop
		}

		Ticks++
		runqueue.update_instances()

		if full_instance.ProcessingState == 1 {
			// Cancel all other runs and return this instance as result.
			current_instance = full_instance
			new_current_instance(current_instance)
			base.Message.Printf("(TODO) *** All constraints OK @ %d ***\n", Ticks)
			break
		} else {
			p := full_instance.TtData.Progress
			if p > full_progress {
				full_progress = p
				full_progress_last = Ticks
				base.Message.Printf(
					"(TODO) [%d] ? %s (%d @ %d)\n",
					Ticks,
					full_instance.TtData.Description,
					full_progress,
					full_progress_last,
				)
			}
		}

		if Ticks == TIMEOUT {
			base.Message.Printf(
				"(TODO) [%d] TIMEOUT (%d @ %d)\n",
				Ticks,
				full_progress,
				full_progress_last,
			)
			break
		}

		//TODO: Special treatment if there are no constraints to add?

		if stage == 0 {
			// During stage 0 only `full_instance` and `null_instance`
			// (and perhaps hard_only_instance) are running.
			switch null_instance.ProcessingState {
			case 0:
				if null_instance.TtData.Ticks == null_instance.Timeout {
					abort_instance(null_instance)
				}
			case 1:
				// The null instance completed successfully.
				current_instance = null_instance
				new_current_instance(current_instance)
				// Start trials of single constraint types.
				constraint_list, _ = get_basic_constraints(
					null_instance, 0)
				// Queue instances for running
				for _, bc := range constraint_list {
					runqueue.add(bc)
				}
				//TODO: This is the initial set, possibly excluding some ...
				base.Message.Printf("(TODO) [%d] INITIAL CONSTRAINT-TYPES: %d\n",
					Ticks, len(constraint_list))
				stage = 1
			default:
				// The null instance failed.
				stage = 10
				base.Message.Printf(
					"(TODO) [%d] Unconstrained instance failed", Ticks)

				base.Error.Println(" ... " + tt_data.Message)

				//TODO: Seek problems in the unconstrained data.
				panic("TODO")
			}
			continue
		}

		if stage <= 2 {
			// During stage 1 we are accumulating single-constraint-type
			// instances: check their states.
			// In stage 2 the constraints which failed (or weren't included)
			// in stage 1 are tried again with longer timeouts. This stage
			// runs until there are no more failing constraints or, more
			// likely, the overall timeout is reached.
			next_timeout := 0

			// See if an instance has completed successfully.
			for i, instance := range constraint_list {
				if instance.ProcessingState == 1 {
					// Completed successfully, make this instance the new base.
					current_instance = instance
					new_current_instance(current_instance)
					next_timeout = max(
						instance.TtData.Ticks*NEW_BASE_TIMEOUT_FACTOR/10,
						STAGE_TIMEOUT)
					// Remove it from constraint list.
					constraint_list = slices.Delete(
						constraint_list, i, i+1)

					break
				}
			}
			if len(constraint_list) == 0 {
				// ... all constraint trials finished.
				if stage == 1 {
					base.Message.Printf(
						"(TODO) [%d] Stage 1 ended\n", Ticks)
					stage = 2
				}
				// Start trials of remaining (hard) constraints.
				STAGE_TIMEOUT = max(STAGE_TIMEOUT,
					current_instance.TtData.Ticks) *
					NEW_STAGE_TIMEOUT_FACTOR / 10
				var n int
				constraint_list, n = get_basic_constraints(
					current_instance, stage)
				if len(constraint_list) == 0 {
					base.Message.Printf(
						"(TODO) [%d] Stage 2 ended\n", Ticks)
					stage = 3
				} else {
					base.Message.Printf(
						"(TODO) [%d] Remaining: %d (timeout %d)\n",
						Ticks, n, STAGE_TIMEOUT)
					// Queue instances for running
					for _, bc := range constraint_list {
						runqueue.add(bc)
					}
				}
				continue
			}

			// Seek failed instances, which should be split.
			// If there is a new base, stop the old instances and
			// restart them accordingly.
			split_instances := []*TtInstance{}
			new_constraint_list := []*TtInstance{}
			for _, instance := range constraint_list {
				if instance.ProcessingState == 2 {
					// failed (or timed out): split the instance

					// Split if more than one instance in list
					if len(instance.Constraints) > 1 {
						timeout := next_timeout
						if timeout == 0 {
							timeout = instance.Timeout
						}
						nhalf := len(instance.Constraints) / 2
						split_instances = append(split_instances,
							new_instance(
								current_instance,
								instance.TtData.Description,
								instance.ConstraintType,
								instance.Constraints[:nhalf],
								timeout))
						split_instances = append(split_instances,
							new_instance(
								current_instance,
								instance.TtData.Description,
								instance.ConstraintType,
								instance.Constraints[nhalf:],
								timeout))
					} else {
						if len(instance.Constraints) != 1 {
							panic("Bug, expected a single constraint")
						}
						// Add to sidelined instances, for next restart
						sidelined = append(sidelined, instance)
					}
				} else {
					if next_timeout != 0 {
						// Cancel existing instance
						if instance.ProcessingState == 0 {
							abort_instance(instance)
						}
						// Indicate that a queued instance is not to be started
						instance.ProcessingState = 3
						// Build new instance
						instance = new_instance(
							current_instance,
							instance.TtData.Description,
							instance.ConstraintType,
							instance.Constraints,
							next_timeout)
						runqueue.add(instance)
					}
					new_constraint_list = append(
						new_constraint_list, instance)
				}
			}
			constraint_list = append(new_constraint_list,
				split_instances...)
			if next_timeout != 0 {
				for _, instance := range sidelined {
					// Build new instance
					instance = new_instance(
						current_instance,
						instance.TtData.Description,
						instance.ConstraintType,
						instance.Constraints,
						next_timeout)
					runqueue.add(instance)
					new_constraint_list = append(
						new_constraint_list, instance)
				}
				sidelined = []*TtInstance{}
			}

			for _, instance := range split_instances {
				runqueue.add(instance)
			}
			continue
		}

		//TODO: A "stuck" analysis on running instances might help to
		// reduce processing time?

	} // tickloop: end

	//TODO: Consider also the possibility that there may be no (or only one)
	// basic constraint types.

	result := current_instance
	ttdata := result.TtData

	nn := 0
	nall := 0
	for i, clist := range result.HardConstraintEnabled {
		n := 0
		for _, b := range clist {
			if b {
				n++
			}
		}
		fmt.Printf("$ CONSTRAINT %d: %d / %d (%s)\n",
			i, n, len(clist), timetable.ConstraintType(i).String())
		nn += n
		nall += len(clist)
	}
	fmt.Printf("$ ALL CONSTRAINTS: %d / %d\n", nn, nall)

	//TODO
	base.Message.Printf("(TODO) RESULT: %s\n", ttdata.Description)
}

func abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		timetable.BACKEND.Abort(instance.TtData)
		instance.Stopped = true
	}
}

func new_instance(
	instance_0 *TtInstance,
	descriptor string,
	constraint_type timetable.ConstraintType,
	constraint_indexes []int,
	timeout int,
) *TtInstance {
	// Prepare instnace "name"
	InstanceCounter++
	if i := strings.Index(descriptor, "~"); i >= 0 {
		descriptor = descriptor[i+1:]
	}
	descriptor = fmt.Sprintf("z%05d~%s", InstanceCounter, descriptor)
	// Copy original TtData (shallow copy only!)
	ttdata := new_ttdata(instance_0.TtData, descriptor)

	// Make a deep copy of the hard constraint matrix
	hcmat0 := instance_0.HardConstraintEnabled
	hcmat := make([][]bool, len(hcmat0))
	for i, c := range hcmat0 {
		hcmat[i] = slices.Clone(c)
	}

	// Make a new `TtInstance`
	instance := &TtInstance{
		Timeout:               timeout,
		TtData:                ttdata,
		HardConstraintEnabled: hcmat,

		// Base data for this instance:
		BaseInstance:   instance_0,
		ConstraintType: constraint_type,
		Constraints:    constraint_indexes,

		// Run time
		Stopped: false,
	}

	// Enable the constraints in `ttdata` and `hcmat`
	//fmt.Printf("§ENABLE %s: %v\n", constraint_type.String(), constraint_indexes)

	// Mark the constraints in the matrix
	cmap := instance.HardConstraintEnabled[constraint_type]
	for _, i := range constraint_indexes {
		cmap[i] = true
	}
	// Reconstruct the constraint list
	newlist := []any{}
	for i, c := range TtData_0.HardConstraints[constraint_type] {
		if cmap[i] {
			newlist = append(newlist, c)
		}
	}
	instance.TtData.HardConstraints[constraint_type] = newlist

	return instance
}

func new_ttdata(
	ttdata_0 *timetable.TtData,
	descriptor string,
) *timetable.TtData {
	ttdata := &timetable.TtData{
		Description:             descriptor,
		SharedData:              ttdata_0.SharedData,
		TeacherNotAvailable:     ttdata_0.TeacherNotAvailable,
		ClassNotAvailable:       ttdata_0.ClassNotAvailable,
		RoomNotAvailable:        ttdata_0.RoomNotAvailable,
		WITHOUT_ROOM_PLACEMENTS: ttdata_0.WITHOUT_ROOM_PLACEMENTS,
		State:                   0,
	}

	// Make a deeper copy of the constraints so that these can be
	// switched on or off without affecting those in the original `TtData`.

	// Make a copy of the constraints lists
	hcmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.HardConstraints))
	for k, v := range ttdata_0.HardConstraints {
		hcmap[k] = slices.Clone(v)
	}
	ttdata.HardConstraints = hcmap

	scmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.SoftConstraints))
	for k, v := range ttdata_0.SoftConstraints {
		scmap[k] = slices.Clone(v)
	}
	ttdata.SoftConstraints = scmap
	return ttdata
}

/*
// This will probably need some tuning. There is probably no such thing as
// an optimal algorithm, that depends very much on the data.
func timed_out(instance *TtInstance) bool {
	if instance.Timeout < 0 {
		// This allows a complete override of the timeout feature.
		// Unlike Timeout = 0 it will not try to catch runs which get
		// stuck at low completion levels.
		return false
	}
	if instance.Timeout != 0 && instance.Ticks > instance.Timeout {
		return true
	}
	delta := instance.Ticks - instance.LastTime
	if delta < 5 {
		return false
	}
	// The acceptable time should depend on Progress.
	if instance.Progress < 80 && delta*2 > instance.Progress {
		return true
	}
	if instance.Progress < 95 && delta > 400 {
		return true
	}
	return false
}
*/

//
/*
// TODO:  Is this old bit fetching file names from db.ModuleData still
// needed somehow?

	fetfile := stempath
	mapfile := stempath
	thisdir := filepath.Dir(stempath)
	moduleData := db.ModuleData
	fetData, ok := moduleData["FetData"].(map[string]string)
	if ok {
		var f string
		f, ok = fetData["FetFile"]
		if ok {
			fetfile = filepath.Join(thisdir, f)
		}
		f, ok = fetData["MapFile"]
		if ok {
			mapfile = filepath.Join(thisdir, f)
		}
	}
	fetfile += ".fet"
	mapfile += ".map"
*/
