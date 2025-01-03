# Druckoptionen

Manche Eigenschaften der Ausdrucke werden in den Typst-Skripten festgelegt. Andere können über die JSON-Datei geregelt werden, insbesondere über die PrintTables-Objekte.

## Die PrintTables-Objekte

Viele der Optionen werden über die "Typst"-Eigenschaft direkt an das Typst-Skript übergeben. Die möglichen Felder hängen also vom eingesetzten Typst-Skript ab. Das Typst-Skript sollte idealerweise sinnvolle Voreinstellungen für so viele Felder wie möglich haben.

Welche Pläne erstellt werden, wird durch die Objekte in der "PrintTables"-Liste festgelegt. Jedes Objekt beschreibt ein Stundenplandokument. Folgende Dokumenttypen sind vorgesehen:

 - Gesamtplan Klassen (eine Klasse pro Seite)
 - Übersichtsplan Klassen
 - Gesamtplan Lehrer (eine Lehrkraft pro Seite)
 - Übersichtsplan Lehrer
 - Gesamtplan Räume (ein Raum pro Seite)
 - Übersichtsplan Räume

Ob ein Übersichtsplan oder ein Gesamtplan erzeugt wird, hängt vom verwendeten Typst-Skript ("TypstTemplate") ab – sie werden von den gleichen Daten konstruiert.

Auch Einzelpläne können erstellt werden, indem die Element-ID des entsprechenden Objekts (Klasse, Lehrer oder Raum) als "Type" angegeben wird.

```
  "PrintTables": [
    {
      "Type": "Class", // oder "Teacher" oder "Room" oder Element-Id
      "TypstTemplate": "template1",
      "TypstJson": "timetable",
      "Pdf": "timetable"

      "Typst": {
        "Title": "Stundenpläne der Klassen",
        "Subtitle": "Entwurf Erstes Halbjahr",
        "PageHeading": "Klasse: %S",

        "WithTimes": true,
        "WithBreaks": true,
        "FieldPlacement": {
          "c": "SUBJECT",
          "tl": "TEACHER",
          "tr": "GROUP",
          "bl": "",
          "br": "ROOM"
        },
        "LastChange": "12.04.2024 um 8:30 Uhr",
        "Legend": {
          "Remark": "Eine Anmerkung",
          "Subjects": [["EU", "Eurythmie"]],
          "Teachers": [["HM", "Hans Müller"], ["MM", "Mara Musterfrau"]],
          "Rooms": []
        },
      },

      "Pages": [
        {
          "Id": Element-Id,
          "LastChange": "18.04.2024 um 18:30 Uhr",
          "Legend": {
            "Remark": "Meine Anmerkung",
            "Subjects": [["EU", "Eurythmie"]],
            "Teachers": [["HM", "Hans Müller"], ["MM", "Mara Musterfrau"]],
            "Rooms": []
          }
        },
            ...
      ]
    },
      ...
  ],
```

### Das Typst-Objekt

Die Eigenschaften dieser Objekte können unabhängig von W365toTypst-Programm gestaltet werden. Sie werden unverändert an das Typst-Skript weitergegeben.

Bei "PageHeading" gibt es über "%N" und "%S" die Möglichkeit Vollnamen und Kurznamen der jeweiligen Klasse, usw., einzubinden.

Über die Option "WithTimes" kann die Zeitangabe ein- bzw. ausgeschaltet werden. Anhand der Option "WithBreaks" wird entschieden, ob nur die Unterrichtsstunden oder auch die Pausen in der Tabelle dargestellt werden. Damit diese funktionieren können, müssen die "Hours" korrekte "Start"- und "End"- Werte haben.

### Die Pages-Objekte

Die Eigenschaften dieser Objekte werden (außer "Id") in die entsprechenden "Pages"-Objekte der Typst-JSON-Eingabe-Datei (siehe unten). Sie ermöglichen individuelle Anpassungen der einzelnen Seite eines Gesamtplans.

## Die Typst-JSON-Eingabe-Datei

Die Daten werden an das Typst-Skript als JSON-Datei mit folgender Struktur übergeben:

```
{
  "TableType": "Class",
  "Info": {
    "Institution": "Musterschule Mulmingen",
    "Days": [
      {
        "Name": "Montag",
        "Short": "Mo"
      },
      ...
    ],
    "Hours": [
      {
        "Name": "1. Stunde",
        "Short": "(1)",
        "Start": "07:35",
        "End": "08:25"
      },
      ...
    ]
  },
  "Typst": {
    ... // von PrintTable
  },
  "Pages": [
    {
      "Name": "1. Klasse",
      "Short": "1",
      "LastChange": "18.04.2024 um 18:30 Uhr",
      "Legend": {
        "Remark": "Meine Anmerkung",
        "Subjects": [["EU", "Eurythmie"]],
        "Teachers": [["HM", "Hans Müller"], ["MM", "Mara Musterfrau"]],
        "Rooms": []
      }
      "Activities": [
        {
          Day:      0,
          Hour:     4,
          Duration: 2,
          Subject:  "Ch",
          Groups:   ["10"],
          Teachers: ["AT"]
          //Rooms:    [],
          //Fraction: 1,
          //Offset:   0,
          //Total:    1,
          //Background: "#FFFFFF",
          //Footnote: "Meine andere Anmerkung"
        },
        ...
      ]
    },
    ...
  ]
}
```

Der Name der Institution sollte im W365TT-Objekt, Feld "Institution", zur Verfügung stehen.
