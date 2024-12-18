# Die Constraint-Elemente

Alle Felder sollten Werte haben, Voreinstellungen gibt es nur in Ausnahmefällen.

Im Allgemeinen kann ein Kurs ein Course-Element oder ein SuperCourse-Element sein.

## LessonsEndDay

Die Lessons des Kurses sollten am Ende des Schülertags liegen.

```
{
    "constraint" :  "MARGIN_HOUR",
    "weight" :      73,
    "course" :      "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
}
```

## BeforeAfterHour

Die Lessons der Kurse sollten vor ("after": false) bzw. nach ("after": true)der angegebenen Stunde – ausschließlich dieser Stunde – liegen.

```
{
	"constraint":   "BEFORE_AFTER_HOUR",
	"weight":       100,
	"courses":      [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"after":        false,
	"hour":         4
}
```

## AutomaticDifferentDays

Die Lessons eines Kurses sollen an unterschiedlichen Tagen stattfinden. Mit „"consecutiveIfSameDay": true“ sollten sie – falls sie doch am selben Tag sind – direkt nacheinander sein.

Dieser Constraint wird im Prinzip auf alle Kurse (mit zwei oder mehr Lessons) angewendet. Wenn dieser Constraint nicht vorhanden ist, wird er mit "weight": 100 angewendet.

Einzelne Kurse können durch DaysBetween-Constraints anders geregelt werden.

```
{
	"constraint":           "AUTOMATIC_DIFFERENT_DAYS",
	"weight":               100,
	"consecutiveIfSameDay": true
}
```

## DaysBetween

Dieser Constraint ist wie AutomaticDifferentDays, erlaubt aber andere Tagesabstände als 1 und wird auf einzelne Kurse angewendet. Mit „"daysBetween": 1“ wird der globale Constraint AutomaticDifferentDays für diese Kurse ausgesetzt.

```
{
	"constraint":           "DAYS_BETWEEN",
	"weight":               100,
	"courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"daysBetween":          2.
	"consecutiveIfSameDay": true
}
```

## DaysBetweenJoin

Anders als DaysBetween wird dieser Constraint zwischen den einzelnen Stunden zweier verschiedener Kurse angewendet, also Kurs 1, Lesson1 : Kurs 2, Lesson 1; Kurs 1, Lesson 2 : Kurs 2, Lesson 2; Kurs 1, Lesson2 : Kurs 2, Lesson 1; ...

```
{
	"constraint":           "DAYS_BETWEEN_JOIN",
	"weight":               100,
	"course1":              "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"course2":              "5fda67de-bbb3-48a2-a098-d957796b7743",
	"daysBetween":          1,
	"consecutiveIfSameDay": false
}
```

## ParallelCourses

Die Lessons der Kurse sollen gleichzeitig stattfinden. Die Anzahl und Länge der Lessons müssen in allen Kursen gleich sein.

```
{
	"constraint":           "PARALLEL_COURSES",
	"weight":               100,
	"courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
}
```

**TODO**: Folgende Constraints sind noch nicht klar. Es kann gut sein, dass sie auf bestimmte Klassen (...) begrenzt oder anders formuliert sein sollten.

## DoubleLessonNotOverBreaks

Dieser Constraint sollte höchstens einmal vorkommen. Eine Doppelstunde soll nicht durch eine Pause unterbrochen werden. Die Pause sind unmittelbar vor den angegebenen Stunden.

```
{
	"constraint":           "DOUBLE_LESSON_NOT_OVER_BREAKS",
	"weight":               90,
	"hours":                [2, 4]
}
```

## NotOnSameDay

```
{
	"constraint":           "NOT_ON_SAME_DAY",
	"weight":               90,
	"subjects":             [
        "8c3b3b63-a51e-4b52-aaa6-d8fadfe6d099",
        "d6e584cb-2409-4d9a-8786-b13978a75aba"
    ]
}
```

## MinHoursFollowing

```
{
	"constraint":   "MIN_HOURS_FOLLOWING",
	"weight":       90,
	"course1":      "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"course2":      "5fda67de-bbb3-48a2-a098-d957796b7743",
	"hours":        4
}
```
