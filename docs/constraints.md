# Constraints

Some constraints are implicit (such as those preventing collisions). Others are specified along with the items they constrain, for example blocked times or number of gaps per day for teachers or classes. Others are specified in the top-level Constraints list. These often constrain the lessons of one or more courses, perhaps the relationships between lessons of different courses. The currently supported constraint types in the Constraints list are described below.

## The Constraints List

All constraints have a Weight field, which is an integer between 0 and 100. A value of 0 means the constraint is inactive, a value of 100 specifies a "hard" constraint.

### AutomaticDifferentDays

The lessons of a course should all be on different days.

This constraint may be specified at most once. If it is not present, a hard constraint affecting all courses is implemented. Note, however, that this constraint can be overridden for individual courses by DaysBetween constraints (see below).

 - Weight (integer)
 - ConsecutiveIfSameDay (boolean): In cases where the constraint is not fulfilled, the lessons should be consecutive. This is only relevant if the constraint is soft.

### DaysBetween

The lessons of a course should have at least the given number of days between them (where a value of 1 is equivalent to different-days).

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.
 - DaysBetween (integer)
 - ConsecutiveIfSameDay (boolean): If the lessons do end up on the same day, they should be consecutive.

If DaysBetween is 1, this constraint will override AutomaticDifferentDays for the listed courses. Otherwise it is a distinct constraint.

### DaysBetweenJoin

This constraint applies between the individual lessons of the two courses, not between the lessons of a course itself. That is, between course 1, lesson 1 and course 2 lesson 1; between course 1, lesson 1 and course 2, lesson 2, etc. Otherwise it is similar to the DaysBetween constraint.

 - Weight (integer)
 - Course1 (Course or SuperCourse reference)
 - Course2 (Course or SuperCourse reference)
 - DaysBetween (integer)
 - ConsecutiveIfSameDay (boolean): If the lessons do end up on the same day, they should be consecutive.

### LessonsEndDay

The lessons of the specified course should be the last lessons of the day for the student group concerned.

 - Weight (integer)
 - Course (Course or SuperCourse reference)

### BeforeAfterHour

The lessons of the specified courses should lie before (or after) the given hour (excluding the given hour).

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.
 - After (boolean): If true, then after, if false, then before.
 - Hour (integer)
