# Waldorf 365: Schnittstelle für die Stundenplanung

## Ausgabeformat

Die Daten werden als JSON-Objekt ausgegeben, der Dateiname sollte mit "_w365.json" enden.

### Top-Level-Objekt

```
{
    "W365TT": {},
    "PrintTables": [],
    "FetData": {},
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
    "Lessons": [],
    "Constraints": {}
}
```

Die Array-Werte enthalten die ggf. geordneten Elemente des entsprechenden Typs. Alle Elemente sind JSON-Objekte. Diese Elemente haben ein optionales „Type“-Feld, das den Namen des Elements ("Day", "Hour", usw.) enthält.

Einige Element-Namen sind anders als die entsprechenden Waldorf-365-Elemente:

 - „TimedObject“ -> „Hour“
 - „Grade“ -> „Class“
 - „GradePartiton“ [sic!] -> „Division“ (im Top-Level-Objekt nicht vorhanden, da kein Top-Level-Element)
 - „EpochPlanCourse“ -> „SuperCourse“

Neu sind „W365TT“, „PrintTables“, „FetData“, „RoomGroup“ und „Constraint“. Es gibt auch  „SubCourse“ – einen Epochenkurs –, das als Unterelement von "SuperCourse" auftaucht.

#### W365TT

In diesem Objekt könnten allgemeine Informationen oder Eigenschaften, die nirgendwo anders richtig passen, erscheinen, z.B.:

```
  "W365TT": {
    "Institution": "Musterschule Mulmingen",
    "Scenario": "96138a85-d78f-4bd0-a5a7-bc8debe29320",
    "FirstAfternoonHour":   6,
    "MiddayBreak":          [5, 6, 7]
  },
```

"schoolName" ist z.B. für Ausdrucke nützlich. "firstAfternoonHour" und "middayBreak" sind für Constraints notwendig, die Nachmittagsstunden und Mittagspausen regeln.

#### PrintTables

Eine Liste PrintTable-Objekte. Jedes dieser Objekt enthält Parameter für den Ausdruck (PDF-Ausgabe über Typst) eines Stundenplans (verschiedene Darstellungen sind möglich). Für die Stundenplanung selbst ist diese Liste nicht relevant. Die Felder werden in einem anderen Dokument beschrieben: [Druckoptionen](druckoptionen.md#druckoptionen).

#### FetData

Ein optionales Objekt, das die Erstellung der FET-Datei konfigurieren kann.
Die möglichen Eigenschaften sind:

```
    "FetFile": string (name of FET file, without ".fet" ending)
    "MapFile": string (name of Id-map file, without ".map" ending)
```

#### Constraint

Diese Objekte haben verschiedene Felder, aber alle haben:

```
    "Constraint": string (Constraint-Typ)
    "Weight": int (0 – 100)
```

Für diese Objekte gibt es eine eigene Dokumentation: [Die Constraint-Elemente](constraintelemente.md#die-constraint-elemente).

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
	"End":      "08:25"
}
```

"Start" und "End" können auch Zeiten mit Sekunden (z.B. "07:35:00") sein, die Sekunden werden ignoriert. Diese Felder sind ggf. für Constraints notwendig, die Pausenzeiten oder Stundenlängen berücksichtigen, aber vor allem für die Ausdrucke (siehe unter PrintTables die Felder "WithTimes" und "WithBreaks").


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

In Waldorf 365 hat ein Kurs eine Liste „PreferredRooms“. Von dieser Liste muss eins dieser Room-Elemente für jede Stunde (Lesson-Element) des Kurses zur Verfügung stehen. Die einzelnen Stunden können unterschiedliche Räume haben. Diese Liste kann auch leer sein. Alternativ kann die Liste aus *einer* RoomGroup-Referenz bestehen. Dann braucht jede Stunde alle Räume der Raumgruppe (sie sind „Pflichträume“).

In Waldorf 365 wird eine Raumgruppe über das RoomGroup-Feld der Room-Elemente definiert. Da sie eigentlich ganz andere Objekte sind, haben sie hier einen eigenen Element-Typ.

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

"Rooms" enthält nur Room-Referenzen (also real vorhandene Einzelräume).

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
            "Id": "9137860d-a656-400f-b4c1-d3e90cf5a4d8",
            "Type": "Division",
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

"Shortcut“ ist eigentlich nur "Level" + "Letter", aber in dieser Form oft nützlicher.

In Waldorf 365 ist eine Division ein Top-Level-Objekt. Deswegen haben sie ein "Id"-Feld. Da sie nur hier gebraucht werden, erscheinen die Division-Elemente nur im "Divisions"-Feld der Klassen.

#### Group

```
{
    "Id":           "00c1ec1b-5f65-43c1-9a73-5fff8d8751e2",
    "Type":         "Group",
	"Shortcut":     "F"
}
```

#### Course

Für die eigentliche Stundenplanung braucht man nur Lesson-Elemente. Die Kurselemente können aber wichtige Informationen über manche Beziehungen verdeutlichen. Anders als in Waldorf 365 steht hier ein Course-Element nur für einen „normalen“ Kurs. Die Epochenschienen werden durch die SuperCourse-Elements zusammen mit den SubCourse-Elementen (für die Epochenkurse) abgedeckt.

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

Waldorf 365 unterstützt Kurse mit mehr als einem Fach. Nur deswegen ist hier "Subjects" eine Liste. Die "Groups" können Group- oder Class-Elemente sein. Für "PreferredRooms" siehe die "RoomGroup"-Beschreibung.

#### SuperCourse

```
{
    "Id":           "kNWJ5jArzE_hQ9FSl6pE3",
    "Type":         "SuperCourse",
    "EpochPlan":    "271baf6f-151b-4354-b50c-add01622cb10",
	"Subject":      "5791c199-3fa3-4aea-8124-bec9d4a7759e",
    "SubCourses":   [
        SubCourse-Element,
        SubCourse-Element
    ]
}
```

**SubCourse**

Ein SubCourse-Element ist fast genau wie ein Course-Element, darf aber nicht als "Course" eines Lesson-Elements vorkommen. Da die SubCourse-Element nur im Zusammenhang mit einem SuperCourse-Element relevant sind, tauchen sie nur in dessen "SubCourses"-Feld auf. Auch aus Gründen, die mit wiederholten Id-Feldern zu tun haben, ist ein SubCourse kein Top-Level-Element.

```
{
    "Id":           "c0f5c633-534a-43f5-9541-df3d93b771a9",
    "Type":         "SubCourse",
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
    ],
    "Flags":    ["SubstitutionService"],
    "Background": "#FFE080",
    "Footnote": "Eine Anmerkung"
}
```

"Course" kann ein Course- oder ein SuperCourse-Element sein. "LocalRooms" sind die Room-Elemente (nur reale Räume), die dem Lesson-Element zugeordnet sind. Sie sollten kompatibel mit den "PreferredRooms" des Kurses sein.

Wenn ein "Lesson" den Eintrag "SubstitutionService" in der "Flags"-Array hat, wird die Klasse (falls vorhanden) des entsprechenden Kurs ("Course"-Eigenschaft) als „Vertretungsplan-Klasse“ ausgesondert. 

Ein nicht platziertes Lesson-Element hätte:

```
	"Day":          -1,
	"Hour":         -1,
    "Fixed":        false,
    "LocalRooms":   [],
```

"Background" (Hintergrundfarbe) und "Footnote" sind für die Ausdrucke relevant. Die Farbe wird in der Form "#RRGGBB" erwartet. Wenn keine angegeben ist, wird die Voreinstellung im Typst-Skript benutzt.
