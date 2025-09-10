# Automation of Constraint Testing

When constructing more complicated timetables, it can happen that the computation time is uncomfortably long. Although this is a perfectly normal aspect of the process, it can make it difficult to trace problem areas. The users may well be left in a state of uncertainty as to whether a timetable is at all possible with the given data set. How long should they wait before deciding the timetable generation is not going to work? And then ... what steps need to be taken to attain a timetable that might be at least be a starting point?

## The aims

 - To deliver some sort of timetable in a limited time, if necessary by disabling some of the constraints.

 - To identify constraints which are difficult (slowing the timetable generation significantly) or impossible to fulfil. Note that it might be difficult to distinguish between these two cases.

 - To permit, on demand, indefinite generation times, with the possibility of manual interruption.

## Automation

A common approach to troubleshooting is to disable and re-enable groups of constraints, aiming to find those constraints that are difficult (or impossible) to fulfil by narrowing down the areas in which they can lie.

With few constraints the generation of solutions will often be very quick, allowing many tests to be carried out in a short time. Automation can be very helpful at this stage. It won't be able to replace the insights of an expert timetable constructor, but can speed up the process even for them. As the run times get longer the savings will probably be less obvious, but some automation of the process can still assist less experienced timetable constructors. The possibility of running tests in parallel on suitable computer systems might also be able to offer speed gains.

Automation along these lines should be seen as a valuable assistant, but not as a panacea. In general, it can't tell you exactly what needs changing, it can only point to areas where changes might be necessary. Often there will be several – perhaps seemingly unrelated – areas in which changes in the specifications could lead to timetable improvements. Experience and analytical expertise are still very desirable qualities in a timetable constructor.

## The main constraints

First it is necessary to identify the constraints which can be sensibly enabled and disabled automatically, rating them according to their importance. I will not include the specification of the  activities (lessons, etc.) here, but assume these are fixed as regards teachers, student groups, duration and numbers. Under some circumstances it may indeed be necessary to make alterations to these, but this is not in the immediate scope of the automation being considered here.

### Blocked time slots

It is possible to specify time slots which are not available for activities – for each teacher, class and room. These should be regarded as non-negotiable, as one of the primary constraints. This is not only because these constraints are supposed to cover requirements resulting from a school's political and organisational structures, but well-chosen placement restrictions can also make the generation of a timetable much quicker. Nevertheless, checks should be made to ensure that these restrictions are not such that no timetable can be generated.

### Fixed placements

Some activities must take place at particular times. This is also regarded as non-negotiable, but must be checked for viability. Placement conflicts with other fixed activities or with blocked time slots need to be reported and corrected before any further processing can take place.

The first test, therefore, is that all activities can be placed in accordance with the constraints "blocked time slots" and "fixed placements". If this is not the case (or if this takes unexpectedly long), a set of tests on these constraints must be made, for example class-by-class placement, to find where the problems lie. Failing fixed placements should be easy to diagnose as these are placed individually anyway.

The optimal order of constraint tests probably can't be determined in general, it depends too much on the nature of the data set. To ease experimentation, the switches should be kept as simple and modular as possible.

### Teacher-specific constraints

 - Minimum activities per day
 - Maximum activities per day
 - Maximum days
 - Maximum gaps per day
 - Maximum gaps per week: only effective if this is smaller than max-gaps-per-day * days (taking max-days into account)
 - Maximum number of afternoons: the start time of an "afternoon" is specified globally
 - Lunch break: this is a simple switch, whether a break should be forced within a group of time slots on every day – the range of slots is specified globally)

### Class-specific constraints

 - Minimum activities per day
 - Maximum activities per day
 - Maximum gaps per day
 - Maximum gaps per week: only effective if this is smaller than max-gaps-per-day * days
 - Maximum number of afternoons: the start time of an "afternoon" is specified globally
 - Lunch break: this is a simple switch, whether a break should be forced within a group of time slots on every day – the range of slots is specified globally)
 - Force first hour: whether the first time slot of each day must be filled

### Room specifications

Because of the possibility of specifying a choice of possible rooms for an activity, handling the room allocation can be quite difficult. The best way to handle progressive testing is probably to eliminate room specifications for the initial tests, adding these only when the other basic constraints have been checked. Also, only the "necessary" rooms would be tested, the choice lists being left to the end, as it is difficult to allocate from these lists before all the activities have been placed.

### Further constraints

Most other constraints can also come in "soft" form. These are initially not so relevant for the automation, which deals primarily with "hard" constraints. The discussion of the following constraints refers, in general, only to the hard forms.

**Different days / Days between**

It is possible to specify that the individual activities of a group of activities should be on different days, or that a number of days should lie between the activities (where the value "1" means "not on the same day"). The soft version of this constraint can also specify that – should a pair of activities be placed on the same day – they will be consecutive (hard constraint). Because of this latter option, the soft versions of these constraints should not be excluded from the tests.

Ill-chosen use of these constraints can easily lead to difficult or impossible timetables, so good feedback concerning problematic cases should be striven for.

**Before or After a given hour**

This specifies that the activities of a course should be before or after the specified hour, not including the specified hour. This is a relatively simple constraint.

**Parallel courses**

This specifies that the activities of the specified courses should lie in the same time slots. The activities to be parallel must of course have the same duration and number. If this is a hard constraint, it may be possible to optimize its implementation (e.g. by building compound activities).

**Double lesson not over breaks**

This constraint should be specified at most once. The breaks are immediately before the specified hours.

**Minimum hours gap between activities of two courses**

This constraint specifies that between the end of an activity of the first course and the start of one of the second, there should be at least the specified number of time slots. This constraint is ordered, allowing another constraint to specify the gap in the other direction.

**Lessons end day**

Because the end of a school day is often somewhat flexible, this is a difficult constraint to implement. Its resolution may well have to be postponed until all activities have been placed. In some (not always easy to define) cases, a more efficient implementation could be devised, but the interactions with other constraints may not be obvious. In general, if such lessons can be given fixed placements, the overall efficiency could benefit greatly.

## Performing the tests

At the centre of the procedure is a sub-process which runs FET on the data with the currently enabled constraints. As the duration of this process cannot be known, there must be some way of stopping it before a result (success or failure) has been returned. In order to achieve a better result in a given time, the possibility of running several processes simultaneously can be helpful, although it won't be a linear gain because it is in the nature of many of the tests that they are actually sequential. Pre-empting a result (assuming an outcome before it has actually been reached) can speed some sequences up, but if the assumption is wrong, the pre-emptive processing will have been in vain.

Because of the halting problem, it can be helpful to estimate the chances of success. In cases where no progress is made for a long (how long?) time, it might be sensible to assume failure, especially if the state is far from completion. There could even be two stages of failure prediction: at the first stage, the failure branch is activated, at the second the instance is halted and the success branch terminated.

A test step entails running the back-end (timetable generator) with a certain set of constraints enabled. If that test succeeds before timing out, those constraints will be marked as "OK". If the test fails, the set of constraints will be divided in two parts, which are then tested sequentially.

By triggering the failure path before the test has completed, work on the (potential) next step can be started earlier than otherwise, possibly enabling some sort of "less constrained" timetable to be achieved more quickly, though of course this helps only if the test really doesn't succeed. It might be sensible to wait a little before starting the failure path, so that the test has a chance to succeed quickly.

When a test succeeds, the failure path (if it has already started) needs to be cancelled. If a test fails before the failure path has been started, the failure path can be started immediately. Note that even if a partial success on a failure path has been registered as "OK" constraints, this is not a problem if this path is aborted later because a test succeeds, because the constraints on the failure path are always just a subset of those in the main test.

### The first tests

Initially, all constraints are enabled and the back-end is started. Soon afterwards, the failure path is started, dividing up the constraints. This can quickly lead to a multitude of processes, which may lead to reduced performance on a less powerful processor with only a few cores. Thus it is probably sensible to limit the number of simultaneous processes. This can, however, lead to a situation where the time limit is reached before the constraints have been reduced to a stage where generation is quick enough, i.e. no result.

To avoid this saturation problem, it might help to start an additional process, with all constraints disabled. This would – assuming it runs successfully – normally be quick, so no great strain on the system, and provide a minimal complete result, though it would probably be a pretty terrible timetable. If this test failed – or ran very slowly – there would not be much sense in running a more constrained data set, so that could be cancelled. Instead, an analysis of the basic data could be undertaken (e.g. seeking difficult classes or teachers).

The minimally constrained test would retain only the blocked time slots for teachers and classes. Room placements would be suppressed. Assuming this completes quickly, it might be worth adding the (fixed) room placements to check for basic problems in that area.

Unblocking time-slots is probably not an option as these blocks are supposed to be non-negotiable. It is assumed to be up to the user to loosen these constraints, if appropriate, on the basis of the diagnostic reports.

### Ordering the tests

Experience shows that certain constraint types are more likely than others to cause problems. Such constraints concern for instance gaps or "days-between-activities". Some care with handling these might help the process overall.
