# Activity Groups

In this experimental branch I would like to test the idea of adjusting the placement functions to use `ActivityGroup` items rather than using the `Activity` items directly. I hope this can unify handling of things like double lessons and parallel courses, while managing the placement of the lessons of a course in a more coordinated manner.

An `ActivityGroup` has a list of `ActivityLessons`. Let's try first with a model that supports only single lessons. Double lessons would then comprise two single lessons and have a rule to keep them together.

There is also a list of possible placements, a possible placement being a tuple of time-slots, one for each `ActivityLesson`. It would be possible to give each tuple a penalty, if this helps the placement algorithm. The list is generated after placement of the fixed activities, but before any non-fixed activities are placed. Thus the list is a static maximal set, assuming no non-fixed activities have been placed. No slot combinations outside this list need be tried because there will be some hard-blockage.

**Note:** This might also be a good opportunity to test artificially extended days, i.e. far more hours per days than are actually needed. The hope is that this might – at some cost in memory – simplify the handling of "days-between" constraints. Just the hour gap would need to be tested, no division into days would be necessary for the tests. With nhours timetable slots in a day, "placement-days" would need at least (2 * nhours - 1) slots, the last (nhours - 1) being unusable. The minimum distance (between start-times) to ensure d days distance would be nhours + (d - 1) * (2 * nhours - 1).

**TODO**
**Use of ttbase:**

```
	ttinfo := ttbase.MakeTtInfo(db)
	ttinfo.PrepareCoreData()
```
