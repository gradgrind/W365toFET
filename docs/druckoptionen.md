# Druckoptionen

Manche Eigenschaften der Ausdrucke werden in den Typst-Skripten festgelegt. Andere können über die JSON-Datei geregelt werden, insbesondere über das PrintOptions-Objekt.

für die Überschriften gibt es folgende Optionen:
```
    "title": "Hauptüberschrift",
    "subtitle": "Entwurf Erstes Halbjahr | Letzte Änderung 15.06.2020 19:30 Uhr",
    "pageHeadingClass": "Klasse: %S",
    "pageHeadingTeacher": "%N (%S)",
    "pageHeadingRoom": "Raum: %N (%S)",
```

Der Name der Schule steht im W365TT-Objekt, Feld "institution" zur Verfügung.

Im „pageHeadingXXX“ gibt es über „%N“ und „%S“ die Möglichkeit Vollnamen und Kurznamen der jeweiligen Klasse, usw., einzubinden.

Die Gestaltung der Tabellen kann durch folgende Optionen angepasst werden:

```
    "withTimes": false,
    "withBreaks": false,
    "boxTextClass": {
        "c": "SUBJECT",
        "tl": "TEACHER",
        "tr": "GROUP",
        "bl": "-",
        "br": "ROOM"
    },
    "boxTextTeacher": {
        "c": "GROUP",
        "tl": "SUBJECT",
        "tr": "TEACHER",
        "bl": "-",
        "br": "ROOM"
    }
```

Über die Option „withTimes“ gibt es die Möglichkeit die Zeitangabe ein- bzw. auszuschalten. Über die Option „withBreaks“ wird entschieden, ob nur die Unterrichtsstunden oder auch die Pausen in der Tabelle dargestellt werden. Damit diese funktionieren können, müssen die „Hours“ korrekte „Start“- und „End“- Werte haben.

Die Stundenbezeichnungen sind aktuell die Kürzel, die Tag-Bezeichnungen die Namen. **TODO**: Vielleicht sollte man in beiden Fällen Kürzel oder Namen wählen können? Oder die Bezeichnungen in Waldorf 365 festlegen und als Optionen übergeben?

Welche Pläne erstellt werden, wird durch die Option „printTables“ festgelegt, z.B.:

```
    "printTables": ["TEACHER", "CLASS", "ROOM"],
```

Ein Einzelplan kann erstellt werden, indem das Id des entsprechenden Objekts (Klasse, Lehrer oder Raum) angegeben wird:

```
    "printId": "9e3251d6-0ab3-4c25-ab66-426d1c339d37",
```

Wenn dieses Feld leer ist, werden alle Pläne erstellt.
