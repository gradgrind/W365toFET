# Setting up the database

The database is represented internally by the structure `base.DbTopLevel`. It can be easily loaded from other forms, for example JSON, but there are some useful supporting data structures which can be derived from the basic data, and don't need to be stored externally. The first of these is a method on the database itself, `PrepareDb`.

## base.DbTopLevel.PrepareDb

This checks the validity of the `Info.MiddayBreak`, if present, and sorts it.

The `SubCourse`s are initially independent elements with references to their owning `SuperCourse`s. The `SuperCourse`s are here given a list of references to their constituent `SubCourse`s.

Also the `Lesson`s are initially independent elements with references to their containing `Course`s and `SuperCourse`s. The course elements are here given a list of references to their constituent `Lesson`s.

Finally the information about classes and groups is checked and linked, as the groups are initially supplied without references to their classes. These links can be added to the `Group` elements by going through the classes and divisions.

Note that a class normally has a reference to a special group for the whole class in its "ClassGroup" field. But this field can also be empty, signifying a special sort of "class" without groups, which can be used for various purposes (e.g. for stand-in / cover lessons).

## Basic data for timetable processing

For timetable processing further structures are set up by the "ttbase" package, especially the `TtInfo` `struct`, which is constructed by the function `ttbase.MakeTtInfo`. The method `TtInfo.PrepareCoreData` sets up further structures for dealing with timetabling. See the documentation for the package "ttbase".
