# Reading the JSON file from Waldorf 365

The structure of the input file (reflected in the structures in "w365tt/structures.go") is quite close to that of the structures in "base/structures.go", but there are some differences. The functions in package "w365tt" perform the operations necessary to read the data in the input file into these internal structures.

The function `ReadJSON` in "w365tt/fromjson.go" reads the input file into the structures in "w365tt/structures.go". It is called from `LoadJSON`, which then performs the processing necessary to transfer the data into the structures in "base/structures.go". `LoadJSON` is called from "cmd/W365toFET/main.go".

Keeping the `struct`s in "w365tt/structures.go" separate from the internal ones in "base/structures.go" allows some flexibility in the naming and contents of the input data. By adding tags like ``` `json:"Shortcut"` ``` to a field descriptor, the JSON field name can be changed without changing any internal data or code other than this tag.

Some of the structures need little or no translation, so they can be more or less directly copied. There is a suite of methods on `base.DbTopLevel` whose names start with "New" (`NewDay`, `NewHour`, etc.), defined in "base/db_manager.go", which call the method `addElement` to add an entry to the `Elements` mapping. This allows easy access to the element's internal `struct` via the element's `Id` (the `Id` fields from the Waldorf 365 input are retained for this purpose). If new elements (not in the input data) are created, new `Id`s are created for them.

## Blocked time-slots and maximum-afternoons constraints

When reading the data for elements with `MaxAfternoons` constraints (`Class` and `Teacher`) the `handleZeroAfternoons` method is called to block all afternoons (via `NotAvailable`) when the value is 0, the `MaxAfternoons` then being set in that case to -1 (constraint disabled). This method is also used to convert/build the `NotAvailable` structure /list of `TimeSlot`s) for the rooms, even though these have no `MaxAfternoons` constraint.

## Classes

When reading the classes from the input file, some checks on the groups and class-divisions are made, to ensure that they meet with expectations. Also a new `Group` element is created for the whole class. This is significant for the courses which concern whole classes. In Waldorf 365 these would reference a `Class` element rather than a `Group` element. The use of such different types as target can be a bit troublesome, so these are converted to reference the newly created `Group` elements for whole classes.

### `GroupRefMap`

To simplify the handling of the class/group references in the courses of the Waldorf 365 data, this mapping (a field of the `w365tt.DbTopLevel` `struct`) is constructed. It maps Waldorf 365 `Group` references to themselves and Waldorf 365 `Class` references to references to the newly created `Group` elements.

## Rooms

The handling of rooms can get a bit complicated. Waldorf 365 supports "real" rooms (`struct` `Room`) and groups of real rooms (`struct` `RoomGroup`). The latter is basically a list of real rooms and is used by a course which requires more than one room. A course specifies a list of rooms in the field "PreferredRooms". If there is more than one room in this list, a choice is implied (one of the given rooms should be chosen).

**TODO:** Check this! The translator supports a list containing just a single room-group or a list with any number of real rooms.

After reading the real rooms in method `readRooms` and the room-groups in method `readRoomGroups`, the method `checkRoomGroups` performs some basic validity checks on the data and adds tags for any room-groups that may not have one. It also extends the names with a list representation of the tags of the constituent real rooms.

In the "base" package there is also a `RoomChoiceGroup` element, representing a list of real rooms, one of which should be chosen. These are created while reading the courses.

Some additional fields in the `w365tt.DbTopLevel` `struct` are maintained to support the translation process to the structures in the "base package:

 - **RealRooms** maps `Id`s to their `base.Room` element.
 - **RoomTags** maps room "tags" (in Waldorf 365 the "Shortcut") to their `Id`. The "base" structures require that all rooms (not just the real ones, but also room-groups and room-choice-groups) have a tag. The input from Waldorf 365 may contain room-groups without a tag ("Shortcut"). These are initially accepted (in method `readRoomGroups`), but the later call to method `checkRoomGroups` will generate new tags for them. The newly created `RoomChoiceGroup` elements will get their tags in the method `makeRoomChoiceGroup`, which is called when handling the courses.
 - **RoomChoiceNames** maps the names of the room-choice-groups (which are generated from the tags of the constituent real rooms) to their `Id`. It is used to enable reuse of `RoomChoiceGroup` elements when a "PreferredRooms" list (in a course) contains the same rooms that were used by another course.

## Courses

There are three types of course:

 - **Course**: This is a "normal" course. It is associated with a number of "resources" (student groups, teachers and rooms) and a number of lessons.
 - **SuperCourse**: This is a special sort of course which is associated with a number of lessons, but the resources are specified by a number of other special courses ("sub-courses") which specify resources, but no lessons.
 - **SubCourse**: This is a course which may be taught for a limited period of time in the lessons of a "super-course". Alternatively, the combination of super-courses and their sub-courses could sometimes be a convenient way to specify that courses should be taught at the same time.

In the input from Waldorf 365 the sub-courses are supplied as entries in the "SubCourses" field of the `SuperCourse` elements. This is done in this way because of how Waldorf 365 handles these items internally. Important to note here is that their `Id`s may be shared with `Course` elements. Thus they cannot appear as top-level objects (which must have unique `Id`s). In the "base" structures, sub-courses are top-level objects, so they are given new `Id`s to avoid clashes with `Course` elements. It is also possible that a sub-course appears in more than one super-course. In this case it is assumed that the two sub-courses really are identical and only one element is created for the "base" structures. The `base.SubCourse` elements also have a "SuperCourses" field, which is a list of references to the associated `SuperCourse` elements.

The validity of the lists of teachers and student groups (for courses and sub-courses) is checked, the student groups being looked up in `GroupRefMap` to ensure that for whole-class groups the `Group` and not the `Class` element is found.

### Subjects

The "base" structures expect a course to have a single subject. In Waldorf 365 a course may have multiple subjects. The method `getCourseSubject` converts a list of subjects to a single (newly created, "composite") subject. This new subject will have a tag constructed from the tags of the constituent subjects. If the same subject list is used in another course, the created subject will be reused.

To support the translation process there are a couple of additional fields in the `w365tt.DbTopLevel` `struct`:

 - **SubjectMap** maps the subject `Id`s (*not* including those of newly created subjects) to their `base.Subject` element. It is used in the construction of new, composite subjects.
 - **SubjectTags** maps subject tags (including those of newly created subjects) to their `Id`. It is used to enable the reuse of newly created subjects.

In the input from Waldorf 365 the `SuperCourse` items have no "Subject" field, but they do have an "EpochPlan" field. The subject must be taken from the linked `EpochPlan` `struct`. These items are not needed for anything else. If the `EpochPlan` tags are not already defined as subjects, new subjects will be created with them.

### Rooms

In the "base" structures, the courses have a "Room" field, whereas the courses (and sub-courses) in the data from Waldorf 365 have a "PreferredRooms" field. "PreferredRooms" can be a single `RoomGroup` or a list of `Room`s. If there is a list of Rooms (more than one room), this is converted to a `RoomChoiceGroup` (in method `makeRoomChoiceGroup`), so that in the end there should be a single item, a `Room`, `RoomChoiceGroup` or `RoomGroup`. The "Room" field of the "base" course references one of these. Repeated use of the same room list will result in only one `RoomChoiceGroup` being generated, which is then shared.
