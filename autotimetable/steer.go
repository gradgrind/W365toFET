package autotimetable

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/timetable"
	"os"
	"runtime"
	"time"
)

// TODO: It may well be desirable to be able to override this
var MAXPROCESSES int = runtime.NumCPU() //TODO: use this?

/*
A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine.

Each instance can be given a set of "child" functions determining how the
tests proceed. For each of these functions a delay is specified before it
starts (after starting the parent function). It is also possible to specify
a function to be called on "success" of the parent function, and one for
failure. These can also be started pre-emptively by specifying a delay.

A timetable instance can be cancelled, including all child instances. When
an instance finishes, any pre-emptively started child instances on the
branch which is now known to be wrong (success or failure) will be cancelled
automatically, including all their children. Because a run can continue
for a long time with no result, there is also the possibility of stopping
the instance and taking the "failure" branch.

The main steering function starts a run with the fully constrained data and
then enters an event loop which is triggered every second. This monitors the
progress of each active instance and handles the actions resulting from the
specified delays.
*/

var Descriptions map[string]string = map[string]string{
	"COMPLETE":           "All constraints active",
	"ONLY_BLOCKED_SLOTS": "All constraints – except blocked slots – disabled",
}

// TODO: At present this only supports a FET back-end. Perhaps a choice should
// be possible ...

func SteerGeneration(tt_data_0 *timetable.TtData, workingdir string) {

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.

	// Open communication channels
	stop := make(chan bool)
	make_instance := make(chan *timetable.TtInstance)

	{
		// Provide an empty working directory.
		os.RemoveAll(workingdir)
		err := os.Mkdir(workingdir, 0755)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		// First run: all constraints enabled, no timeout
		//TODO: On successful completion, all other instances should be stopped.
		// If it fails, just this instance should be wound up. Otherwise it
		// should still be running when the whole process finishes, and would
		// need stopping.
		instance := &timetable.TtInstance{
			//Id:          0,
			Description: "COMPLETE",
			Ticks:       0,
			WorkingDir:  workingdir,
			TtData_0:    tt_data_0,
			TtData:      tt_data_0,
			NewInstance: make_instance,
			Stop:        stop,

			FailurePath: timetable.TtChainedFunc{
				Delay: 1, Func: test_sequence},
		}

		instance.FailurePath = timetable.TtChainedFunc{
			Delay: 1, Func: test_sequence}

		// Request start of instance
		make_instance <- instance
	}

	// *** Channel reader loop ***

	/*
	 * Aborting an instance will lead to its demise in its own goroutine,
	 * probably with a state resulting from the interruption (-2). The
	 * consequences might be picked up here only on the next tick. It is
	 * also vaguely possible that the instance completes naturally, so it
	 * is important to avoid unexpected consequences. It may be safest to
	 * suppress any changes resulting from the completion.
	 */

	active_instances := map[*timetable.TtInstance]struct{}{}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	ended := []*timetable.TtInstance{}
	for {
		ended = ended[:0]
		select {
		case <-stop:
			//fmt.Println("checkProgress done!")
			//TODO: tidying up?
			return
		case new_instance := <-make_instance:
			//TODO: The timetable "engine" should be replaceable.
			fet.NewFet(new_instance)
			active_instances[new_instance] = struct{}{}
		case <-ticker.C:
			// Update the progress records of the currently active
			// subprocesses, handle tick-related events.
			for inst := range active_instances {
				// Once `inst.State` is no longer 0, the instance is removed
				// from the active list. Its children are only started if they
				// derive from a successful run or an unsuccessful one that
				// was not terminated by `cancelAll`.
				// `inst.State` is changed only once.
				if inst.State != 0 {
					ended = append(ended, inst)
					// If appropriate, activate follow-on processes.
					if inst.State == 1 {
						// successful ...
						cancelAll(inst.FailureInstance)
						if inst.SuccessPath.Delay >= 0 {
							inst.SuccessInstance = inst.SuccessPath.Func(inst)
							inst.SuccessPath.Delay = -1
						}
					} else if inst.State != -3 {
						// failed ...
						cancelAll(inst.SuccessInstance)
						if inst.FailurePath.Delay >= 0 {
							inst.FailureInstance = inst.FailurePath.Func(inst)
							inst.FailurePath.Delay = -1
						}
					}
					continue
				}
				inst.Ticks++
				h := inst.UpdateHandler
				if h != nil {
					// The handler should only be set when the the process is
					// fully running
					h(inst)
					if inst.State != 0 {
						continue
					}
				}

				// Handle timeout
				if inst.Ticks == inst.Timeout {
					inst.State = -2
					inst.Abort(inst.HandlerData)
					continue
				}

				// Handle starting of follow-on paths
				if inst.SuccessPath.Delay == inst.Ticks {
					inst.SuccessInstance = inst.SuccessPath.Func(inst)
					inst.SuccessPath.Delay = -1 // flag already started
				}
				if inst.FailurePath.Delay == inst.Ticks {
					inst.FailureInstance = inst.FailurePath.Func(inst)
					inst.FailurePath.Delay = -1 // flag already started
				}
				for _, chfunc := range inst.OtherPaths {
					if chfunc.Delay == inst.Ticks {
						inst.OtherInstances = append(inst.OtherInstances,
							chfunc.Func(inst))
						//chfunc.Delay = -1
						//inst.OtherPaths[i] = chfunc
					}
				}
			}
			// Remove terminated instances from the active list
			for _, inst := range ended {
				delete(active_instances, inst)
			}
		}
	}
}

// Stop an instance and all of its children. Not only the "success" branch is
// to be stopped, but also all the others.
func cancelAll(instance *timetable.TtInstance) {
	if instance == nil {
		return
	}
	if instance.State == 0 {
		instance.Abort(instance.HandlerData)
		instance.State = -3
	}
	cancelAll(instance.FailureInstance)
	cancelAll(instance.SuccessInstance)
	for _, i := range instance.OtherInstances {
		cancelAll(i)
	}
}

func test_sequence(instance_0 *timetable.TtInstance) *timetable.TtInstance {
	// Copy original DbTopLevel (shallow copy only!)
	db0 := instance_0.TtData_0.Db
	db := *db0

	// Copy original TtData (shallow copy only!)
	tt_data := *instance_0.TtData_0
	tt_data.Db = &db

	// Make a new `TtInstance`
	instance := &timetable.TtInstance{
		Description: "ONLY_BLOCKED_SLOTS",
		Ticks:       0,
		WorkingDir:  instance_0.WorkingDir,
		TtData_0:    instance_0.TtData_0,
		TtData:      &tt_data,
		NewInstance: instance_0.NewInstance,
		Stop:        instance_0.Stop,
		//TODO: follow-on paths
	}

	// Keep only the hard-blocked time slots and the fixed activities.

	// Remove activity constraints
	tt_data.Constraints = map[string][]any{}
	tt_data.MinDaysBetweenLessons = nil
	tt_data.ParallelLessons = nil
	tt_data.WITHOUT_ROOM_PLACEMENTS = true

	// Regenerate the teachers list without constraints
	new_teachers := make([]*base.Teacher, len(db.Teachers))
	for i, t0p := range db.Teachers {
		t := *t0p
		t.MinLessonsPerDay = -1 // unconstrained
		t.MaxLessonsPerDay = -1 // unconstrained
		t.MaxDays = -1          // unconstrained
		t.MaxGapsPerDay = -1    // unconstrained
		t.MaxGapsPerWeek = -1   // unconstrained
		t.MaxAfternoons = -1    // unconstrained
		t.LunchBreak = false
		new_teachers[i] = &t
	}
	db.Teachers = new_teachers

	// Regenerate the classes list without constraints
	new_classes := make([]*base.Class, len(db.Classes))
	for i, c0p := range db.Classes {
		c := *c0p
		c.MinLessonsPerDay = -1 // unconstrained
		c.MaxLessonsPerDay = -1 // unconstrained
		c.MaxGapsPerDay = -1    // unconstrained
		c.MaxGapsPerWeek = -1   // unconstrained
		c.MaxAfternoons = -1    // unconstrained
		c.LunchBreak = false
		c.ForceFirstHour = false
		new_classes[i] = &c
	}
	db.Classes = new_classes

	// Request start of instance
	instance.NewInstance <- instance
	return instance
}

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
