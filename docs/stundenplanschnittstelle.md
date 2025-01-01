# Waldorf 365: Schnittstelle für die Stundenplanung

## Ausgabeformat

Die Daten werden als JSON-Objekt ausgegeben, der Dateiname sollte mit "_w365.json" enden.

### Top-Level-Objekt

```
{
    "w365TT": {},
    "printTables": [],
    "days": [],
    "hours": [],
    "teachers": [],
    "subjects": [],
    "rooms": [],
    "roomGroups": [],
    "classes": [],
    "groups": [],
    "courses": [],
    "superCourses": [],
    "lessons": [],
    "constraints": {}
}
```

Die Array-Werte enthalten die ggf. geordneten Elemente des entsprechenden Typs. Alle Elemente sind JSON-Objekte. Diese Elemente haben ein optionales „Type“-Feld, das den Namen des Elements ("Day", "Hour", usw.) enthält. Diese Namen werden großgeschrieben, als Objekteigenschaften aber kleingeschrieben.

Einige Element-Namen sind anders als die entsprechenden Waldorf-365-Elemente:

 - „TimedObject“ -> „Hour“
 - „Grade“ -> „Class“
 - „GradePartiton“ [sic!] -> „Division“ (im Top-Level-Objekt nicht vorhanden, da kein Top-Level-Element)
 - „EpochPlanCourse“ -> „SuperCourse“

Neu sind „W365TT“, „PrintTables“, „RoomGroup“ und „Constraint“. Es gibt auch  „SubCourse“ – einen Epochenkurs –, das als Unterelement von "SuperCourse" auftaucht.

#### W365TT

In diesem Objekt könnten allgemeine Informationen oder Eigenschaften, die nirgendwo anders richtig passen, erscheinen, z.B.:

```
  "w365TT": {
    "institution": "Musterschule Mulmingen",
    "scenario": "96138a85-d78f-4bd0-a5a7-bc8debe29320",
    "firstAfternoonHour":   6,
    "middayBreak":          [5, 6, 7]
  },
```

"schoolName" ist z.B. für Ausdrucke nützlich. "firstAfternoonHour" und "middayBreak" sind für Constraints notwendig, die Nachmittagsstunden und Mittagspausen regeln.

#### PrintTables

Eine Liste PrintTable-Objekte. Jedes dieser Objekt enthält Parameter für den Ausdruck (PDF-Ausgabe über Typst) eines Stundenplans (verschiedene Darstellungen sind möglich). Für die Stundenplanung selbst ist diese Liste nicht relevant. Die Felder werden in einem anderen Dokument beschrieben: [Druckoptionen](druckoptionen.md#druckoptionen).

#### Constraint

Diese Objekte haben verschiedene Felder, aber alle haben:

```
    "constraint": string (Constraint-Typ)
    "weight": int (0 – 100)
```

Für diese Objekte gibt es eine eigene Dokumentation: [Die Constraint-Elemente](constraintelemente.md#die-constraint-elemente).

### Die Top-Level-Elemente

Ich habe in jedem Element ein „Type“-Eigenschaft, das den Typ des Elements angibt. Da die Elementtypen über die Datenstruktur erkennbar sind, sind diese Felder nicht wirklich notwendig und könnten weggelassen werden. Vielleicht helfen sie aber, die JSON-Dateien etwas übersichtlicher zu machen. 

#### Day

```
{
    "id":       "069240ee-b709-4fe2-813a-f04ce8c3614e",
    "type":     "Day",
	"name":     "Montag",
	"shortcut": "Mo"
}
```

#### Hour

```
{
    "id":       "58708f73-4703-4975-b6e2-ebb03971ff5a",
    "type":     "Hour",
    "name":     "Hauptunterricht I",
	"shortcut": "HU I",
    "start":    "07:35",
	"end":      "08:25"
}
```

"start" und "end" können auch Zeiten mit Sekunden (z.B. "07:35:00") sein, die Sekunden werden ignoriert. Diese Felder sind ggf. für Constraints notwendig, die Pausenzeiten oder Stundenlängen berücksichtigen, aber vor allem für die Ausdrucke (siehe unter PrintTables die Felder "WithTimes" und "WithBreaks").


#### Teacher

```
{
    "id":           "9e3251d6-0ab3-4c25-ab66-426d1c339d37",
    "type":         "Teacher",
    "name":         "Bach",
	"shortcut":     "SB",
	"firstname":    "Sebastian",
	"absences":     [
        {"day": 0, "hour": 7},
        {"day": 0, "hour": 8}
    ],
	"minLessonsPerDay": 2,
	"maxLessonsPerDay": -1,
	"maxDays":          -1,
	"maxGapsPerDay":    -1,
    "maxGapsPerWeek":   3,
	"maxAfternoons":    -1,
    "lunchBreak":       true
}
```

Bei den Min-/Max-Constraints bedeutet -1, dass der Constraint nicht aktiv ist.

#### Subject

```
{
    "id":           "5791c199-3fa3-4aea-8124-bec9d4a7759e",
    "type":         "Subject",
    "name":         "Hauptunterricht",
	"shortcut":     "HU"
}
```

#### Room

```
{
    "id":       "f0d7a9e4-841e-4585-adee-38cde49aa676",
    "type":     "Room",
    "name":     "Klassenzimmer 1",
	"shortcut": "k1",
	"absences": []
}
```

#### RoomGroup

In Waldorf 365 hat ein Kurs eine Liste „PreferredRooms“. Von dieser Liste muss eins dieser Room-Elemente für jede Stunde (Lesson-Element) des Kurses zur Verfügung stehen. Die einzelnen Stunden können unterschiedliche Räume haben. Diese Liste kann auch leer sein. Alternativ kann die Liste aus *einer* RoomGroup-Referenz bestehen. Dann braucht jede Stunde alle Räume der Raumgruppe (sie sind „Pflichträume“).

In Waldorf 365 wird eine Raumgruppe über das RoomGroup-Feld der Room-Elemente definiert. Da sie eigentlich ganz andere Objekte sind, haben sie hier einen eigenen Element-Typ.

```
{
    "id":       "f0d7a9e4-841e-4585-adee-38cde49aa676",
    "type":     "RoomGroup",
    "name":     "ad hoc room group for epoch plans w2, k11, auv, ch, mu",
	"shortcut": "adhoc11",
	"rooms":    [
        "36dd1fff-8a20-42f2-a3b6-27b244e10150",
        "541ae8c8-5c31-4e80-8d08-a1935b13294e",
        "63d6f30e-e064-490e-9f95-cd212eb6c435",
        "d4daeb99-d562-4a41-a2ec-ca11cd2d4bca",
        "d84476e9-01fa-4396-a703-cdec8fd2ec13"
    ]
}
```

"rooms" enthält nur Room-Referenzen (also real vorhandene Einzelräume).

#### Class

```
{
    "id":           "5b6cbd2c-d27f-4e73-8a56-a1c7d348b727",
    "type":         "Class",
    "name":         "Klasse 10",
	"shortcut":     "10",
	"absences":     [],
	"level":        10,
    "letter":       "",
    "divisions":    [
        {
            "id": "9137860d-a656-400f-b4c1-d3e90cf5a4d8",
            "type": "Division",
            "name": "A und B Gruppe",
            "groups": [
                "904e8c12-a817-49a1-9fc2-f554a19f5873",
                "4c35be10-2519-41bb-9539-4c4caf95f8e7"
            ]
        },
        {
            "name": "Fremdsprachengruppen",
            "groups": [
                "00c1ec1b-5f65-43c1-9a73-5fff8d8751e2",
                "32c960be-d2da-4cdf-ad80-6ab61f45aef6",
                "cc92e228-52a2-412a-b861-de9952d87a51"
            ]
        },
        {
            "name": "Leistungsgruppen",
            "groups": [
                "b7a88739-e323-49a8-911d-7ba67cb746cd",
                "0fc5740c-1706-475c-b496-ec722e8c5a58"
            ]
        }
    ],
	"minLessonsPerDay": 4,
	"maxLessonsPerDay": -1,
	"maxGapsPerDay":    -1,
    "maxGapsPerWeek":   1,
	"maxAfternoons":    3,
    "lunchBreak":       true,
	"forceFirstHour":   true
}
```

"shortcut“ ist eigentlich nur "level" + "letter", aber in dieser Form oft nützlicher.

In Waldorf 365 ist eine Division ein Top-Level-Objekt. Deswegen haben sie ein "id"-Feld. Da sie nur hier gebraucht werden, erscheinen die Division-Elemente nur im "divisions"-Feld der Klassen.

#### Group

```
{
    "id":           "00c1ec1b-5f65-43c1-9a73-5fff8d8751e2",
    "type":         "Group",
	"shortcut":     "F",
}
```

#### Course

Für die eigentliche Stundenplanung braucht man nur Lesson-Elemente. Die Kurselemente können aber wichtige Informationen über manche Beziehungen verdeutlichen. Anders als in Waldorf 365 steht hier ein Course-Element nur für einen „normalen“ Kurs. Die Epochenschienen werden durch die SuperCourse-Elements zusammen mit den SubCourse-Elementen (für die Epochenkurse) abgedeckt.

```
{
    "id":           "c0f5c633-534a-43f5-9541-df3d93b771a9",
    "type":         "Course",
    "subjects":     [
        "12165a63-6bf9-4b81-b06c-10b141d6c94e"
    ],
	"groups":       [
        "2f6082ce-0eb9-45ff-b2e8-a5475462454c"
    ],
    "teachers":     [
        "f24f0ed6-f5ad-423e-9a6c-6a46536b85ab"
    ],
    "preferredRooms":   [
        "541ae8c8-5c31-4e80-8d08-a1935b13294e",
        "4d44ae7e-0e31-4aa0-a539-d4b2570b1b5c",
        "7d0c09fa-eaf6-4298-8faa-afeb1f4477c4"
    ]
}
```

Waldorf 365 unterstützt Kurse mit mehr als einem Fach. Nur deswegen ist hier "subjects" eine Liste. Die "groups" können Group- oder Class-Elemente sein. Für "preferredRooms" siehe die "RoomGroup"-Beschreibung.

#### SuperCourse

```
{
    "id":           "kNWJ5jArzE_hQ9FSl6pE3",
    "type":         "SuperCourse",
    "epochPlan":    "271baf6f-151b-4354-b50c-add01622cb10",
	"subject":      "5791c199-3fa3-4aea-8124-bec9d4a7759e",
    "subCourses":   [
        SubCourse-Element,
        SubCourse-Element
    ]
}
```

**SubCourse**

Ein SubCourse-Element ist fast genau wie ein Course-Element, darf aber nicht als "course" eines Lesson-Elements vorkommen. Da die SubCourse-Element nur im Zusammenhang mit einem SuperCourse-Element relevant sind, tauchen sie nur in dessen "subCourses"-Feld auf. Auch aus Gründen, die mit wiederholten Id-Feldern zu tun haben, ist ein SubCourse kein Top-Level-Element.

```
{
    "id":           "c0f5c633-534a-43f5-9541-df3d93b771a9",
    "type":         "SubCourse",
    "subjects":     [
        "12165a63-6bf9-4b81-b06c-10b141d6c94e"
    ],
	"groups":       [
        "2f6082ce-0eb9-45ff-b2e8-a5475462454c"
    ],
    "teachers":     [
        "f24f0ed6-f5ad-423e-9a6c-6a46536b85ab"
    ],
    "preferredRooms":   [
        "541ae8c8-5c31-4e80-8d08-a1935b13294e","4d44ae7e-0e31-4aa0-a539-d4b2570b1b5c","7d0c09fa-eaf6-4298-8faa-afeb1f4477c4"
    ]
}
```

#### Lesson

```
{
    "id":       "sSLW2M3LKhxTjMk_MWU_h",
    "type":     "Lesson",
    "course":   "kNWJ5jArzE_hQ9FSl6pE3",
	"duration": 1,
	"day":      0,
	"hour":     0,
    "fixed":    true,
	"localRooms":   [
        "f28f3540-dd02-4c6d-a166-78bb359c1f26"
    ],
    "background": "#FFE080",
    "footnote": "Eine Anmerkung"
}
```

"course" kann ein Course- oder ein SuperCourse-Element sein. "localRooms" sind die Room-Elemente (nur reale Räume), die dem Lesson-Element zugeordnet sind. Sie sollten kompatibel mit den "preferredRooms" des Kurses sein.

Ein nicht platziertes Lesson-Element hätte:

```
	"Day":          -1,
	"Hour":         -1,
    "Fixed":        false,
    "LocalRooms":   []
```

"background" (Hintergrundfarbe) und "footnote" sind für die Ausdrucke relevant. Die Farbe wird in der Form "#RRGGBB" erwartet. Wenn keine angegeben ist, wird die Voreinstellung im Typst-Skript benutzt.
