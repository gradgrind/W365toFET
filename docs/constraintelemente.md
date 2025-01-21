# Die Constraint-Elemente

Alle Felder sollten Werte haben, Voreinstellungen gibt es nur in Ausnahmefällen.

Im Allgemeinen kann ein Kurs ein Course-Element oder ein SuperCourse-Element sein.

## LessonsEndDay

Die Lessons des Kurses sollten am Ende des Schülertags liegen.

```
{
    "Constraint" :  "MARGIN_HOUR",
    "Weight" :      73,
    "Course" :      "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
}
```

## BeforeAfterHour

Die Lessons der Kurse sollten vor ("After": false) bzw. nach ("After": true)der angegebenen Stunde – ausschließlich dieser Stunde – liegen.

```
{
	"Constraint":   "BEFORE_AFTER_HOUR",
	"Weight":       100,
	"Courses":      [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"After":        false,
	"Hour":         4
}
```

## AutomaticDifferentDays

Die Lessons eines Kurses sollen an unterschiedlichen Tagen stattfinden. Mit „"ConsecutiveIfSameDay": true“ sollten sie – falls sie doch am selben Tag sind – direkt nacheinander sein.

Dieser Constraint wird im Prinzip auf alle Kurse (mit zwei oder mehr Lessons) angewendet. Wenn dieser Constraint nicht vorhanden ist, wird er mit "Weight": 100 angewendet.

Einzelne Kurse können durch DaysBetween-Constraints anders geregelt werden.

```
{
	"Constraint":           "AUTOMATIC_DIFFERENT_DAYS",
	"Weight":               100,
	"ConsecutiveIfSameDay": true
}
```

## DaysBetween

Dieser Constraint ist wie AutomaticDifferentDays, erlaubt aber andere Tagesabstände als 1 und wird auf einzelne Kurse angewendet. Mit „"DaysBetween": 1“ wird der globale Constraint AutomaticDifferentDays für diese Kurse ausgesetzt.

```
{
	"Constraint":           "DAYS_BETWEEN",
	"Weight":               100,
	"Courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"DaysBetween":          2.
	"ConsecutiveIfSameDay": true
}
```

## DaysBetweenJoin

Anders als DaysBetween wird dieser Constraint zwischen den einzelnen Stunden zweier verschiedener Kurse angewendet, also Kurs 1, Lesson1 : Kurs 2, Lesson 1; Kurs 1, Lesson 2 : Kurs 2, Lesson 2; Kurs 1, Lesson2 : Kurs 2, Lesson 1; ...

```
{
	"Constraint":           "DAYS_BETWEEN_JOIN",
	"Weight":               100,
	"Course1":              "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"Course2":              "5fda67de-bbb3-48a2-a098-d957796b7743",
	"DaysBetween":          1,
	"ConsecutiveIfSameDay": false
}
```

## ParallelCourses

Die Lessons der Kurse sollen gleichzeitig stattfinden. Die Anzahl und Länge der Lessons müssen in allen Kursen gleich sein.

```
{
	"Constraint":           "PARALLEL_COURSES",
	"Weight":               100,
	"Courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
}
```

**TODO**: Folgende Constraints sind noch nicht klar. Es kann gut sein, dass sie auf bestimmte Klassen (...) begrenzt oder anders formuliert sein sollten.

## DoubleLessonNotOverBreaks

Dieser Constraint sollte höchstens einmal vorkommen. Eine Doppelstunde soll nicht durch eine Pause unterbrochen werden. Die Pause sind unmittelbar vor den angegebenen Stunden.

```
{
	"Constraint":           "DOUBLE_LESSON_NOT_OVER_BREAKS",
	"Weight":               90,
	"Hours":                [2, 4]
}
```

## NotOnSameDay

```
{
	"Constraint":           "NOT_ON_SAME_DAY",
	"Weight":               90,
	"Subjects":             [
        "8c3b3b63-a51e-4b52-aaa6-d8fadfe6d099",
        "d6e584cb-2409-4d9a-8786-b13978a75aba"
    ]
}
```

## MinHoursFollowing

```
{
	"Constraint":   "MIN_HOURS_FOLLOWING",
	"Weight":       90,
	"Course1":      "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"Course2":      "5fda67de-bbb3-48a2-a098-d957796b7743",
	"Hours":        4
}
```
