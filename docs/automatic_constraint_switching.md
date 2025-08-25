# Automatic Constraint Switching

When constructing more complicated timetables, it can happen that the processing time is uncomfortably long. Although this is a perfectly normal aspect of the process, it can make it difficult to trace problem areas.

A common approach to troubleshooting is to disable and re-enable groups of constraints, aiming to find those constraints that are difficult (or impossible) to fulfil by narrowing down the areas in which they can lie.

With few constraints the generation of solutions will often be very quick, allowing many tests to be carried out in a short time. Automation can be very helpful at this stage. It won't be able to replace the insights of an expert timetable constructor, but can speed up the process even for them. As the run times get longer the savings will probably be less obvious, but some automation of the process can still assist less experienced timetable constructors. The possibility of running tests in parallel on suitable computer systems can also speed up the process considerably.

Automation along these lines should be seen as a valuable assistant, but not as a magic bullet. In general, it can't tell you exactly what needs changing, it can only point to areas where changes might be necessary. Often there will be several – perhaps seemingly unrelated – areas in which changes in the specifications could lead to timetable improvements. Experience and analytical expertise are still very desirable qualities in a timetable constructor.

## The main constraints

First it is necessary to identify the constraints which can be sensibly enabled and disabled automatically, rating them according to their importance. I will not include the specification of the  activities (lessons, etc.) here, but assume these are fixed as regards teachers, student groups, duration and numbers. Under some circumstances it may indeed be necessary to make alterations to these, but this is not in the scope of the automation being considered here.

### Blocked time slots

It is possible to specify time slots which are not available for activities – for each teacher, class and room. These should be regarded as non-negotiable, as one of the primary constraints. This is not only because these constraints are supposed to cover requirements resulting from a school's political and organisational structures, but well-chosen placement restrictions can also make the generation of a timetable much quicker. Nevertheless, checks should be made to ensure that these restrictions are not such that no timetable can be generated.

### Fixed placements

Some activities must take place at particular times. This is also regarded as non-negotiable, but must be checked for viability. Placement conflicts with other fixed activities or with blocked time slots need to be reported and corrected before any further processing can take place.

The first test, therefore, is that all activities can be placed in accordance with the constraints "blocked time slots" and "fixed placements". If this is not the case (or if this takes unexpectedly long), a set of tests on these constraints must be made, for example class-by-class placement, to find where the problems lie. Failing fixed placements should be easy to diagnose as these are placed individually anyway.

The optimal order of constraint tests probably can't be determined in general, it depends too much on the nature of the data set. To ease experimentation, the switches should be kept as simple and modular as possible.

## Teacher-specific constraints

 - Minimum activities per day
 - Maximum activities per day
 - Maximum days
 - Maximum gaps per day
 - Maximum gaps per week: only effective if this is smaller than max-gaps-per-day * days (taking max-days into account)
 - Maximum number of afternoons: the start time of an "afternoon" is specified globally
 - Lunch break: this is a simple switch, whether a break should be forced within a group of time slots on every day – the range of slots is specified globally)

## Class-specific constraints

 - Minimum activities per day
 - Maximum activities per day
 - Maximum gaps per day
 - Maximum gaps per week: only effective if this is smaller than max-gaps-per-day * days
 - Maximum number of afternoons: the start time of an "afternoon" is specified globally
 - Lunch break: this is a simple switch, whether a break should be forced within a group of time slots on every day – the range of slots is specified globally)
 - Force first hour: whether the first time slot of each day must be filled


## Room specifications

Because of the possibility of specifying a choice of possible rooms for an activity, handling the room allocation can be quite difficult. The best way to handle progressive testing is probably to eliminate room specifications for the initial tests, adding these only when the other basic constraints have been checked. Also, only the "necessary" rooms would be tested, the choice lists being left to the end, as it is difficult to allocate from these lists before all the activities have been placed.

## Further constraints

Most other constraints can also come in "soft" form. These are initially not so relevant for the automation, which deals primarily with "hard" constraints. The discussion of the following constraints refers, in general, only to the hard forms.

### Different days / Days between

It is possible to specify that the individual activities of a group of activities should be on different days, or that a number of days should lie between the activities (where the value "1" means "not on the same day"). The soft version of this constraint can also specify that – should a pair of activities be placed on the same day – they will be consecutive (hard constraint). Because of this latter option, the soft versions of these constraints should not be excluded from the tests.

Ill-chosen use of these constraints can easily lead to difficult or impossible timetables, so good feedback concerning problematic cases should be striven for.

### Before or After a given hour

This specifies that the activities of a course should be before or after the specified hour, not including the specified hour. This is a relatively simple constraint.

### Parallel courses

This specifies that the activities of the specified courses should lie in the same time slots. The activities to be parallel must of course have the same duration and number. If this is a hard constraint, it may be possible to optimise its implementation (e.g. by building compound activities).

### Double lesson not over breaks

This constraint should be specified at most once. The breaks are immediately before the specified hours.

### Minimum hours gap between activities of two courses

This constraint specifies that between the end of an activity of the first course and the start of one of the second, there should be at least the specified number of time slots. This constraint is ordered, allowing another constraint to specify the gap in the other direction.

### Lessons end day

Because the end of a school day is often somewhat flexible, this is a difficult constraint to implement. Its resolution may well have to be postponed until all activities have been placed. In some (not always easy to define) cases, a more efficient implementation could be devised, but the interactions with other constraints may not be obvious. In general, if such lessons can be given fixed placements, the overall efficiency could benefit greatly.
