# Room Choice Lists and "SuperCourses"

Handling rooms for `SuperCourses` is difficult, because there is in general not enough information available to know which rooms must be available in parallel. A `SubCourse` can specify which room(s) it needs (a `Room`, `RoomGroup` or `RoomChoiceGroup`), but because the timetabler – in the absence of a finished plan for the placement of the `SubCourses` within the year – can't know which `SubCourses` need to be allocated in parallel. Even if this information could be made available somehow, it's use can lead to potentially fragile solutions where the movement of a `SubCourse` could break the room allocations.

Thus it is strongly recommended to use only fixed rooms for `SubCourses`. The fixed room(s) of all the `SubGroups` are taken as fixed for the `SuperCourse`. It makes no difference if a room appears twice (duplicates simply being removed).

The room specification for `SuperCourses` is handled as follows:

1) The rooms specified for all the `SubCourses` are combined, rejecting duplicates, whether they are fixed rooms (`Rooms` and the `Rooms` within `RoomGroups`) or the sets of `Rooms` within `RoomChoiceGroup`. The total number of rooms to be reserved is the number of unique fixed rooms plus the number of unique choice lists (the order of `Rooms` within a `RoomChoiceGroup` is irrelevant).

Thus, in order to ensure that an extra room is reserved, either a new fixed room or a new (distinct from those of the other `SubCourses`) room-choice list must be specified for a `SubCourse`. 

At the end of this step there is a list of fixed rooms and a list of room-choice lists.

2) The fixed rooms are eliminated from the room-choice lists. If a list now contains only one room, this list is removed from the room-choice lists and the remaining room is added to the fixed rooms (i.e. no change in total number of rooms) – and then step 2 is repeated. If a room-choice list is empty, it is removed from the room-choice lists, causing the total number of rooms to drop. If the room total drops below the expectation, an error is reported and an attempt at recovery is made (simply dropping all room-choice lists). Note that the room total can increase at step 4.

3) The room-choice lists are then analysed by constructing their Cartesian product, eliminating result values which contain repeated rooms and duplicate values (room sets – the order of the rooms is irrelevant). With an invalid set of room-choice lists it is possible that this will lead to an empty result list. In that case, the error is reported and an attempt at recovery is made (simply dropping all room-choice lists).

4) The individual value lists of the Cartesian product are then searched for rooms which are present in all value lists. These rooms are added to the fixed rooms (thus increasing the room total). If any changes were made, the process is started again at step 2.

Depending on the algorithm used to "solve" the timetable, it may be more efficient to use a list of the final Cartesian product values than to use the list of room-choice lists.
