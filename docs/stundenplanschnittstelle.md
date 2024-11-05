# Waldorf 365: Schnittstelle für die Stundenplanung

## Ausgabeformat

Aktuell gibt es folgende Elemente, die für die Stundenplanung relevant sind:

 - Day
 - TimedObject
 - Absence
 - Teacher
 - Subject
 - Room
 - Grade
 - Group
 - GradePartiton (sic)
 - Course
 - EpochPlanCourse
 - Lesson
 - Fraction

Grundsätzlich könnten diese so bleiben, wie sie sind. Es müssten nur überflüssige Felder und Referenzen ohne vorhandenes Ziel entfernt werden.

Einige Element-Typen sind allerdings nicht wirklich „Top-Level-Elemente“ mit unabhängiger Existenz – deren Elemente gehören praktisch einem anderen Element. Das sind:

 - Absence (gehört einem Teacher-, Grade- oder Room-Element)
 - GradePartit[i]on (gehört einem Grade-Element)
 - Fraction (gehört einem Lesson-Element)

Die Absence- und GradePartit[i]on-Elemente würde ich, in vereinfachter Form (und ohne Id-Feld) als Eigenschaften der Elemente, denen sie gehören, umgestalten. Die Fraction-Elemente würde ich völlig weglassen, da das Schnittstellenprogramm in einer für sich geeigneten Form aus den GradePartit[i]on-Elementen und den Group-Referenzen der Course-Elemente besser erstellen kann.

Meine Wunschvorstellung für die Datenstruktur würde dann ungefähr wie unten dargestellt aussehen (als JSON-Dokument gedacht).

### Top-Level-Objekt

```
{
    "W365TT": {},
    "Days": [],
    "Hours": [],
    "Teachers": [],
    "Subjects": [],
    "Rooms": [],
    "RoomGroups": [],
    "Classes": [],
    "Groups": [],
    "Courses": [],
    "SuperCourses": [],
    "SubCourses": [],
    "Lessons": [],
    "Constraints": {}
}
```

Ich habe die Mehrzahl für die Schlüssel benutzt, um die Listennatur zu verdeutlichen.

Die Array-Werte enthalten die ggf. geordneten Elemente des entsprechenden Typs. Alle Elemente sind JSON-Objekte.

Einige Namen habe ich geändert, weil sie mir passender erschienen, das ist aber ein unwichtiges Detail:

 - „TimedObject“ -> „Hour“
 - „Grade“ -> „Class“
 - „GradePartiton“ -> „Division“ (hier nicht vorhanden, da kein Top-Level-Element)
 - „EpochPlanCourse“ -> „SuperCourse“

Neu sind hier „W365TT“, „RoomGroup“, „SubCourse“ und „Constraint“.

#### W365TT

In diesem Objekt könnten allgemeine Informationen oder Eigenschaften, die nirgendwo anders richtig passen, erscheinen, z.B.:
```
  "W365TT": {
    "SchoolName": "Musterschule",
    "Scenario": "96138a85-d78f-4bd0-a5a7-bc8debe29320"
  },
```

#### RoomGroups

Statt über das RoomGroup-Feld der Room-Elemente eine Raumgruppe (die eigentlich ganz anders als ein Room ist) zu definieren, finde ich einen eigenen Element-Typ klarer.

#### SubCourse

Für die eigentliche Stundenplanung braucht man nur Lesson-Elemente. Die Kurselemente können aber wichtige Informationen über manche Beziehungen verdeutlichen. Für die „normalen“ Kurse finde ich die Bezeichnung „Course“ passend. Die Epochenschienen (die ich mit „SuperCourse“ bezeichne) haben selbst (eigentlich) keine Lehrer, Gruppen oder Räume. Diese kommen aus den dazugehörigen Epochenkursen. Der Klarheit und Flexibilität wegen hätte ich gern auch diese Epochenkurse („SubCourse“) dabei. Den SubCourse-Elementen sind keine Lesson-Elemente zugeordnet. Ein SubCourse-Element hat aber eine Referenz zum übergeordneten SuperCourse-Element.

#### Constraint

Diese Objekte sind noch zu definieren.

### Die Top-Level-Elemente

Ich habe in jedem Element ein „Type“-Eigenschaft, das den Typ des Elements angibt. Da die Elementtypen über die Datenstruktur erkennbar sind, sind diese Felder nicht wirklich notwendig und könnten weggelassen werden. Vielleicht helfen sie aber, die JSON-Dateien etwas übersichtlicher zu machen. 

#### Day

```
{
    "Id":       "069240ee-b709-4fe2-813a-f04ce8c3614e",
    "Type":     "Day",
	"Name":     "Montag",
	"Shortcut": "Mo"
}
```

#### Hour

```
{
    "Id":       "58708f73-4703-4975-b6e2-ebb03971ff5a",
    "Type":     "Hour",
    "Name":     "Hauptunterricht I",
	"Shortcut": "HU I",
    "Start":    "07:35",
	"End":      "08:25",
	"FirstAfternoonHour":   false,
	"MiddayBreak":          false
}
```

Die Eigenschaften „FirstAfternoonHour“ und „MiddayBreak“ wären vielleicht besser im W365TT-Element, in etwas klarerer Form, untergebracht:

```
    "FirstAfternoonHour":   6,
    "MiddayBreak":          [5, 6, 7],
```


#### Teacher

```
{
    "Id":           "9e3251d6-0ab3-4c25-ab66-426d1c339d37",
    "Type":         "Teacher",
    "Name":         "Bach",
	"Shortcut":     "SB",
	"Firstname":    "Sebastian",
	"Absences":     [
        {"Day": 0, "Hour": 7},
        {"Day": 0, "Hour": 8}
    ],
	"MinLessonsPerDay": 2,
	"MaxLessonsPerDay": -1,
	"MaxDays":          -1,
	"MaxGapsPerDay":    -1,
    "MaxGapsPerWeek":   3,
	"MaxAfternoons":    -1,
    "LunchBreak":       true
}
```

Bei den Min-/Max-Constraints bedeutet -1, dass der Constraint nicht aktiv ist.

#### Subject

```
{
    "Id":           "5791c199-3fa3-4aea-8124-bec9d4a7759e",
    "Type":         "Subject",
    "Name":         "Hauptunterricht",
	"Shortcut":     "HU"
}
```

#### Room

```
{
    "Id":       "f0d7a9e4-841e-4585-adee-38cde49aa676",
    "Type":     "Room",
    "Name":     "Klassenzimmer 1",
	"Shortcut": "k1",
	"Absences": []
}
```

#### RoomGroup

```
{
    "Id":       "f0d7a9e4-841e-4585-adee-38cde49aa676",
    "Type":     "RoomGroup",
    "Name":     "ad hoc room group for epoch plans w2, k11, auv, ch, mu",
	"Shortcut": "adhoc11",
	"Rooms":    [
        "36dd1fff-8a20-42f2-a3b6-27b244e10150",
        "541ae8c8-5c31-4e80-8d08-a1935b13294e",
        "63d6f30e-e064-490e-9f95-cd212eb6c435",
        "d4daeb99-d562-4a41-a2ec-ca11cd2d4bca",
        "d84476e9-01fa-4396-a703-cdec8fd2ec13"
    ]
}
```

Bei den Räumen verstehe ich die Regeln so:

 - Die Room-Elemente stehen für die real vorhandenen Räume.
 - Die Rooms-Eigenschaft eines RoomGroup-Elements enthält nur Room-Referenzen. Diese Räume werden für die Stunde unbedingt gebraucht, sind also Pflichträume.
 - Ein Kurs (Course- oder SubCourse-Element) kann über seine PreferredRooms-Eigenschaft eine Liste Room-Referenzen angeben. Von dieser Liste muss eins dieser Room-Elemente für jede Stunde (Lesson-Element) des Kurses zur Verfügung stehen. Die einzelnen Stunden können unterschiedliche Räume haben. Diese Liste kann auch leer sein. Alternativ kann die Liste aus *einer* RoomGroup-Referenz bestehen. Dann braucht jede Stunde alle Räume der Raumgruppe.

#### Class

```
{
    "Id":           "5b6cbd2c-d27f-4e73-8a56-a1c7d348b727",
    "Type":         "Class",
    "Name":         "Klasse 10",
	"Shortcut":     "10",
	"Absences":     [],
	"Level":        10,
    "Letter":       "",
    "Divisions":    [
        {
            "Name": "A und B Gruppe",
            "Groups": [
                "904e8c12-a817-49a1-9fc2-f554a19f5873",
                "4c35be10-2519-41bb-9539-4c4caf95f8e7"
            ]
        },
        {
            "Name": "Fremdsprachengruppen",
            "Groups": [
                "00c1ec1b-5f65-43c1-9a73-5fff8d8751e2",
                "32c960be-d2da-4cdf-ad80-6ab61f45aef6",
                "cc92e228-52a2-412a-b861-de9952d87a51"
            ]
        },
        {
            "Name": "Leistungsgruppen",
            "Groups": [
                "b7a88739-e323-49a8-911d-7ba67cb746cd",
                "0fc5740c-1706-475c-b496-ec722e8c5a58"
            ]
        }
    ],
	"MinLessonsPerDay": 4,
	"MaxLessonsPerDay": -1,
	"MaxGapsPerDay":    -1,
    "MaxGapsPerWeek":   1,
	"MaxAfternoons":    3,
    "LunchBreak":       true,
	"ForceFirstHour":   true
}
```
Ich habe hier ein „Shortcut“-Eigenschaft hinzugefügt (= Level + Letter) – weil es mir nützlich erscheint.

#### Group

```
{
    "Id":           "00c1ec1b-5f65-43c1-9a73-5fff8d8751e2",
    "Type":         "Group",
	"Shortcut":     "F",
}
```

#### Course

```
{
    "Id":           "c0f5c633-534a-43f5-9541-df3d93b771a9",
    "Type":         "Course",
    "Subjects":     [
        "12165a63-6bf9-4b81-b06c-10b141d6c94e"
    ],
	"Groups":       [
        "2f6082ce-0eb9-45ff-b2e8-a5475462454c"
    ],
    "Teachers":     [
        "f24f0ed6-f5ad-423e-9a6c-6a46536b85ab"
    ],
    "PreferredRooms":   [
        "541ae8c8-5c31-4e80-8d08-a1935b13294e",
        "4d44ae7e-0e31-4aa0-a539-d4b2570b1b5c",
        "7d0c09fa-eaf6-4298-8faa-afeb1f4477c4"
    ]
}
```

Es wäre schön, wenn es immer nur ein Subject gäbe, also:
```
	"Subject":      "12165a63-6bf9-4b81-b06c-10b141d6c94e",
```
Das entspricht aber nicht dem aktuellen Stand von Waldorf 365.

Die Ziele der Group-Werte können Group- oder Class-Elemente sein.

#### SuperCourse

```
{
    "Id":           "kNWJ5jArzE_hQ9FSl6pE3",
    "Type":         "SuperCourse",
	"Subject":      "5791c199-3fa3-4aea-8124-bec9d4a7759e"
}
```

#### SubCourse

Das ist fast genau wie ein Course-Element, aber mit zusätzlicher Eigenschaft „SuperCourse“. Ein SubCourse-Element darf nicht als Ziel des Course-Wertes eines Lesson-Elements vorkommen.

```
{
    "Id":           "c0f5c633-534a-43f5-9541-df3d93b771a9",
    "Type":         "SubCourse",
    "SuperCourse":  "kNWJ5jArzE_hQ9FSl6pE3",
    "Subjects":     [
        "12165a63-6bf9-4b81-b06c-10b141d6c94e"
    ],
	"Groups":       [
        "2f6082ce-0eb9-45ff-b2e8-a5475462454c"
    ],
    "Teachers":     [
        "f24f0ed6-f5ad-423e-9a6c-6a46536b85ab"
    ],
    "PreferredRooms":   [
        "541ae8c8-5c31-4e80-8d08-a1935b13294e","4d44ae7e-0e31-4aa0-a539-d4b2570b1b5c","7d0c09fa-eaf6-4298-8faa-afeb1f4477c4"
    ]
}
```

Auch hier wäre es schön, wenn es immer nur ein Subject gäbe, also:
```
	"Subject":      "12165a63-6bf9-4b81-b06c-10b141d6c94e",
```

#### Lesson

```
{
    "Id":       "sSLW2M3LKhxTjMk_MWU_h",
    "Type":     "Lesson",
    "Course":   "kNWJ5jArzE_hQ9FSl6pE3",
	"Duration": 1,
	"Day":      0,
	"Hour":     0,
    "Fixed":    true,
	"LocalRooms":   [
        "f28f3540-dd02-4c6d-a166-78bb359c1f26"
    ]
}
```

Das Ziel des Course-Wertes kann ein Course- oder ein SuperCourse-Element sein.

Ein nicht platziertes Lesson-Element hätte:

```
	"Day":          -1,
	"Hour":         0,
    "Fixed":        false,
    "LocalRooms":   []
```
