package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
	"slices"
)

type RunQueue struct {
	Queue      []*TtInstance
	Active     map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

func (rq *RunQueue) add(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = append(rq.Queue, instance)
	//base.Message.Printf("(TODO) [%d] Queue %s\n",
	//	Ticks, instance.TtData.Description)
}

func (rq *RunQueue) add_front(instance *TtInstance) {
	instance.ProcessingState = -1 // not started yet
	rq.Queue = slices.Insert(rq.Queue, rq.Next, instance)
	base.Message.Printf("(TODO) [%d] Queue Front %s\n",
		Ticks, instance.TtData.Description)
}

func (rq *RunQueue) disable() {
	rq.MaxRunning = 0 // no new starts possible
}

func (rq *RunQueue) instance_completed(instance *TtInstance, state int) {
	instance.ProcessingState = state
	if state != 2 {
		rq.instance_deactivate(instance)
	}
}

func (rq *RunQueue) instance_deactivate(instance *TtInstance) {
	delete(rq.Active, instance)
}

func (rq *RunQueue) update_instances() {
	// First increment the ticks of active instances.
	for instance := range rq.Active {
		ttdata := instance.TtData
		//base.Message.Printf("(TODO) [%d] ? ACTIVE (%d / %d): %s\n",
		//	Ticks, ttdata.State, instance.ProcessingState, ttdata.Description)
		if ttdata.State != 0 && instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick` below.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		if ttdata.State == 0 {
			ttdata.Ticks++
			// Among other things, update the state:
			timetable.BACKEND.Tick(ttdata)
		} else if instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick`.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		//	}

		//	for instance := range rq.Active {
		//		ttdata := instance.TtData

		//???
		if instance.ProcessingState == 3 {
			// Await completion of the goroutine
			if ttdata.State != 0 {
				rq.instance_deactivate(instance)
			}
			continue
		}

		switch ttdata.State {
		case 0: // running, not finished
			// check for timeout
			if instance.Timeout == ttdata.Ticks {
				base.Message.Printf("(TODO) [%d] TIMEOUT %s @ %d (%d)\n",
					Ticks, ttdata.Description, ttdata.Ticks, ttdata.Progress)
				// Stop instance
				abort_instance(instance)
			}

		case 1: // completed successfully
			if instance.Tagged {
				base.Message.Printf("(TODO) [%d] <<+ %s @ %d\n",
					Ticks, ttdata.Description, ttdata.Ticks)
			} else {
				base.Message.Printf("(TODO) [%d] (<+) %s @ %d\n",
					Ticks, ttdata.Description, ttdata.Ticks)
			}
			// Cancel subsidiary instances
			cancel_instance(instance.Instance1)
			cancel_instance(instance.Instance2)
			instance.Result = instance
			rq.instance_completed(instance, 1)

		default: // completed unsuccessfully
			if instance.ProcessingState != 2 {
				// This is done only the first time round for this instance.
				// Some instances remain active after the primary run has
				// completed, so for subsequent loops this block should be
				// skipped.
				if instance.Tagged {
					base.Message.Printf("(TODO) [%d] <<- %s @ %d\n",
						Ticks, ttdata.Description, ttdata.Ticks)
				} else {
					base.Message.Printf("(TODO) [%d] (<-) %s @ %d\n",
						Ticks, ttdata.Description, ttdata.Ticks)
				}
				// This doesn't deactivate the instance (for state = 2):
				rq.instance_completed(instance, 2)
			}

			//TODO: If it is an actual error, the halves should perhaps still
			// be allowed to run, even though they may not have started yet.

			if instance.Instance1 == nil {
				if instance.Instance2 == nil {
					// No halves
					instance.Result = instance.BaseInstance
					rq.instance_deactivate(instance)
				} else {
					// The 1st half has completed, the 2nd half is now
					// building, possibly based on the result of the 1st half.
					if instance.Instance2.Result != nil {
						instance.Result = instance.Instance2.Result
						rq.instance_deactivate(instance)
					}
					// Otherwise the instance remains active (though not
					// running).
				}
				continue
			}

			// else: instance.Instance1 != nil

			if instance.Instance1.Result != nil {
				// The 1st half has completed.
				// Check that it actually added constraints.
				if instance.Instance1.Result == instance.BaseInstance {
					// 1st half: no constraints added,
					// wait for completion of 2nd half.
					instance.Instance1 = nil
				} else {
					if instance.Instance2.ProcessingState < 0 {
						// The 2nd half hasn't started yet: replace its
						// base by the completed 1st half.
						instance.Instance2.BaseInstance = instance.Instance1.Result
						instance.Instance1 = nil

						// Otherwise wait for the 2nd half to finish before
						// adding it to the 1st half.
					} else if instance.Instance2.Result != nil {
						i2 := instance.Instance2.Result
						if i2 == instance.BaseInstance {
							// 2nd half: no constraints added
							instance.Result = instance.Instance1.Result
							rq.instance_deactivate(instance)
						} else {
							next_instance := new_instance(
								instance.Instance1,
								instance.Instance1.TtData.Description+"+",
								i2.ConstraintType,
								i2.Constraints,
								instance.Timeout)
							instance.Instance2 = next_instance
							instance.Instance1 = nil
							rq.add_front(next_instance)
						}
					}
				}
			}
		}
	}
}

func (rq *RunQueue) update_queue() int {
	// Try to start queued instances
	running := 0
	for i := range rq.Active {
		if i.ProcessingState == 0 || i.ProcessingState == 3 {
			running++
		}
	}
	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Next++

		if instance.ProcessingState < 0 {
			instance.ProcessingState = 0 // indicate started/running
			rq.Active[instance] = struct{}{}
			running++
		} else {
			if instance.ProcessingState != 3 {
				panic("Bug")
			}
			// Cancelled before starting, skip it
			continue
		}

		ttdata := instance.TtData

		// If eligible for binary splitting, queue the two halves
		if len(instance.Constraints) > 1 {
			nhalf := len(instance.Constraints) / 2
			instance.Instance1 = new_instance(
				instance.BaseInstance,
				ttdata.Description+"~0",
				instance.ConstraintType,
				instance.Constraints[:nhalf],
				instance.Timeout)
			rq.add(instance.Instance1)

			instance.Instance2 = new_instance(
				instance.BaseInstance,
				ttdata.Description+"~1",
				instance.ConstraintType,
				instance.Constraints[nhalf:],
				instance.Timeout)
			rq.add(instance.Instance2)
		}

		if instance.Tagged {
			base.Message.Printf("(TODO) [%d] >> %s\n",
				Ticks, ttdata.Description)
		} else {
			base.Message.Printf("(TODO) [%d] (>) %s\n",
				Ticks, ttdata.Description)
		}
		timetable.BACKEND.Run(ttdata, TESTING)
	}
	//TODO--
	//fmt.Printf("$ [%d] Running/Active instances: %d/%d\n",
	//	Ticks, running, len(rq.Active))
	//base.Message.Printf("$ [%d] Running/Active instances: %d/%d\n",
	//	Ticks, running, len(rq.Active))
	return len(rq.Active)
}
