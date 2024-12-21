# Druckoptionen

Manche Eigenschaften der Ausdrucke werden in den Typst-Skripten festgelegt. Andere können über die JSON-Datei geregelt werden, insbesondere über das PrintOptions-Objekt.

Viele der Optionen werden über die "Typst"-Eigenschaft direkt an das Typst-Skript übergeben. Die möglichen Felder hängen also vom eingesetzten Typst-Skript ab. Das Typst-Skript sollte idealerweise sinnvolle Voreinstellungen für so viele Felder wie möglich haben.

Welche Pläne erstellt werden, wird durch das Feld "PrintTables" festgelegt. Einzelpläne können erstellt werden, indem das Id der entsprechenden Objekte (Klasse, Lehrer oder Raum) angegeben wird.

```
"printOptions": {

    "printTables": ["Teacher", "Class", "Room"],

    "typst": {
        "PageHeading": {
            "Class": "Klasse: %S",
            "Teacher": "%N (%S)",
            "Room": "Raum: %N (%S)",
        },
        
        "Subtitle": "Entwurf Erstes Halbjahr | Letzte Änderung 15.06.2020 19:30 Uhr",
        
        "WithTimes": false,
        
        "WithBreaks": false,
        
        "FieldPlacements": {
            "Class": {
                "c": "SUBJECT",
                "tl": "TEACHER",
                "tr": "GROUP",
                //"bl": "",
                "br": "ROOM",
            },
            "Teacher": {
                "c": "GROUP",
                "tl": "SUBJECT",
                "tr": "TEACHER",
                //"bl": "",
                "br": "ROOM",
            },
            "Room": {
                "c": "GROUP",
                "tl": "SUBJECT",
                //"tr": "",
                //"bl": "",
                "br": "TEACHER",
            },
        }
    }
}
```

Bei den "PageHeadings" gibt es über "%N" und "%S" die Möglichkeit Vollnamen und Kurznamen der jeweiligen Klasse, usw., einzubinden.

Über die Option "WithTimes" kann die Zeitangabe ein- bzw. ausgeschaltet werden. Anhand der Option "WithBreaks" wird entschieden, ob nur die Unterrichtsstunden oder auch die Pausen in der Tabelle dargestellt werden. Damit diese funktionieren können, müssen die "Hours" korrekte "Start"- und "End"- Werte haben.

Die Daten werden an das Typst-Skript als JSON-Datei mit folgender Struktur übergeben:

```
{
    "TableType": "Room",
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
        ... // von PrintOptions
    },
    "Pages": [
        {
            "Name": "Chemieraum",
            "Short": "ch",
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
                    //Background: "#FFFFFF"
                },
                ...
            ]
        },
        ...
    ]
}
```

Der Name der Institution sollte im W365TT-Objekt, Feld "institution", zur Verfügung stehen.
