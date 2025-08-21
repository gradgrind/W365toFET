# Room Choice Lists and "SuperCourses"

Handling rooms for `SuperCourses` is difficult, because there is in general not enough information available to know which rooms must be available in parallel. A `SubCourse` can specify which room(s) it needs (a `Room`, `RoomGroup` or `RoomChoiceGroup`), but because the timetabler – in the absence of a finished plan for the placement of the `SubCourses` within the year – can't know which `SubCourses` need to be allocated in parallel. Even if this information could be made available somehow, it's use can lead to potentially fragile solutions where the movement of a `SubCourse` could break the room allocations.

Thus it is strongly recommended to use only fixed rooms for `SubCourses`. The fixed room(s) of all the `SubGroups` are taken as fixed for the `SuperCourse`. It makes no difference if a room appears twice (duplicates simply being removed).

If choice lists are indeed used for `SubCourses`, an attempt will be made to "condense" them:

1) Any `RoomChoiceGroup` containing a fixed room (of the `SuperCourse`) will be ignored, assuming the requirement is fulfilled.

2) Any duplicates (same set of real rooms) will be removed.

3) In the remaining `RoomChoiceGroups` the occurrences of individual (real) rooms will be counted. If any rooms appear more than once, the one which appears most often will be added as a fixed room and those `RoomChoiceGroups` containing it will be removed. This step is repeated until there are no more multiple occurrences.

TODO: Unfortunately, there are combinations which could be handled differently on different runs, which might lead to FET files which are sometimes soluble and sometimes not. I should change that!
 
A more promising approach might be to make every choice list (after removing duplicates) specify an additional room ...

Then something like [[r1, r2], [r1, r3], [r2, r3]] would simplify to fixed [r1, r2, r3], and [[r1, r2], [r1, r3], [r2, r3], [r1, r2, r3]] would be rejected as impossible.
