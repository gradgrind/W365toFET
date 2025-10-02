# Testing Constraints' Difficulty

    An algorithm to find particularly difficult constraints, aiming to place all activities within a given time period at the cost of ignoring certain constraints, when necessary.

## An unproblematic data set

The basic idea is that an "instance" with all its constraints enabled should be run for a specified time (the "time limit"). If it completes successfully within that time, all is well.

## The first test of "simplified" data

To cater for situations in which the data is not soluble, or at least not soluble within the specified time, other instances are run in parallel. To avoid general slow-downs, the number of active (running) instances should not exceed the available number of real processors.

(**TODO**) It might be worth running a second instance (in parallel) in which only the hard constraints are retained. Of course, if there aren't any soft constraints this instance could be skipped.

Then there is the "unconstrained instance", in which all non-essential constraints have been removed. The essential constraints which are here still active are blocked times for teachers and classes (and rooms) and fixed activity placements. This instance is given a certain time to run, a time which is much shorter than the overall time limit.

Should the unconstrained instance not complete successfully in the allotted time, a special phase is entered to try to localise the difficulty (e.g. in which class). The details of this phase still need fleshing out (**TODO**).

If the unconstrained instance completes successfully, which it normally should, the next stage of testing is started, in which the constraint types are tested individually.

## Testing individual constraint types

(**TODO**) Note that at present I am only taking hard constraints into account, as the primary aims are to test the viability of the data and to identify constraints which make a data set difficult or impossible to "solve".

The constraints are divided up according to their type so that there is a list of constraints for each type. For each of these types a new instance is started. An instance object contains (among other things) a reference to the (successfully completed) instance on which it builds – which I refer to as the base instance of the new instance. For these individual-constraint-type instances the base is the unconstrained instance. The base instance contains its own completion time. Each new instance also contains the list of constraints to be added, so that there is enough information to split the trial in various ways, should this be desirable.

These basic trials of the individual constraint types are queued to be run in parallel as much as possible. When one of them completes it is added either to a list of successes or to a list of failures. When all of these trials have been completed, the failed instances are sorted to put the ones that got furthest (highest progress percentage) first. The idea is that the constraint types which cause the fewest difficulties should be added first in the next stage. This potentially allows more of the constraints to be integrated into the proposed result (in the given time) than might be the case without this ordering.

## Gradually extending the set of included constraints

Once the first successful basic trials have completed, the next stage can begin. This is the gradual accumulation of constraints, one type at a time. If there are enough real processors available, this stage can start before the previous stage has completed – at any accumulation step only the next completed basic trial is needed.

For the accumulation steps the behaviour of the instances is more complicated. A simple success/failure result is not enough, the aim is to eliminate individual constraints which make the processing time too long.

### Special behaviour of instances which add constraint lists

The timeout of these instances is handled differently, it doesn't stop the run, but starts a further, subsidiary instance with a only a subset of the added constraints. The hope is that by basing the subsidiary trial on binary division of the constraint list, in conjunction with parallel processing, the time to complete the trial can be reduced. It is not "pure" like a traditional binary search, as the two halves cannot be run independently, the second half takes the result of the first half as its base. The final result of each subsidiary instance contains a possibly shortened constraint list. If the "full" instance completes before the subsidiary one, the subsidiary instance is cancelled – it has turned out to be unnecessary.

Note that the subsidiary instances themselves also exhibit this special behaviour.
(**TODO**) Unfortunately this only works as intended if there are enough real processors. If there aren't, at some stage a subsidiary instance will not be able to start until its parent instance finishes, making the subsidiary instances redundant. A better way might be to start the subsidiary instance soon after the full instance (or simultaneously?) and use the timeout as a real timeout? In this way the subsidiary instance's start would just be delayed, reflecting the reduced capabilities due to having fewer processors? The timeout would need to be significantly longer than the base instance completion time!
Or would a better way be to not run the full version at all, just go straight to the divisions?

If there is a timeout, it should probably be at least twice as long as the base time. The worst-case scenario is perhaps when the first constraint in a list is a problem. This could lead to multiple long trials.
