# Data Exchange Formats

## Waldorf 365 Timetable Interface

Waldorf 365 uses a fairly linear, "flattened", form for its data, there is very little direct structural nesting. Basically the whole database is just a list of "objects", these being key-value mappings. The keys are strings, the values also strings, which can represent numbers (integer or float) and boolean as well as text values. Certain of these textual values have a special meaning – as reference to another object. Because each object has its own identifier (the "Id" field) it is possible for an object to reference any other object(s). Even lists of references are supported, by joining them with commas. Thus more elaborate data structures can be represented, somewhat similarly to the way it is done in a relational database.

Basically this is a very simple and flexible way to represent arbitrarily complicated data structures. However, the lack of an inherent hierarchy can lead to very tangled data structures which are difficult to keep track of, rather like "goto" statements in programming languages.

### Waldorf 365 Output

For dealing with timetables only a fraction of the total database is necessary. An interface to an external program (or an internal subsystem) should not include unnecessary ballast. So those objects which are needed need to be sieved out from the rest. Basically these objects are:

 - Days (the days of the school week)
 - Hours (daily teaching periods)
 - Teachers
 - Classes, and groups within them
 - Rooms
 - Subjects
 - Courses and the lessons which belong to them
 - A multitude of diverse constraints which determine where the lessons may be placed

Internally, in Waldorf 365, each of these items is in principle an object. 
Some of these objects provide information pertinent to only one object belonging to the above types. If more hierarchy is allowed in the data format, it might make sense to represent this information as fields in the corresponding object. Examples would be "Absences" and "GradePartitions".

A first suggestion for an interface format can be found (in German) in [this document](stundenplanschnittstelle.md).
