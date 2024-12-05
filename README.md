# W365toFET

```
W365toFET path/to/sp001_w365.json

    -> path/to/sp001_w365.log
    -> path/to/sp001.fet
    -> path/to/sp001.map
```

Die Stundenplan-Daten werden von der JSON-Datei eingelesen.

Ausgegeben wird eine FET-Datei im selben Ordner. Auch eine Logdatei (mit Fehlermeldungen, usw.) und eine Zuordnungsdatei für die FET-Activities werden erstellt.

## Aktueller Stand (05.12.2024)

Bis auf die „Constraint“-Elemente werden alle Elemente in `docs/stundenplanschnittstelle.md` in einigermaßen entsprechende FET-Strukturen übertragen.

Die weiteren Constraints werden jetzt anfänglich übersetzt.

In dieser Version haben Lehrer und Klassen den gleichen Ansatz für Mittagspausen: Eine der Mittagsstunden muss frei sein.

Es wurde bisher nur wenig getestet!

In dieser Version werden die Daten in eine etwas andere interne Struktur gebracht – W365-unabhängig – bevor sie übersetzt werden, siehe „base-package“. Auch unabhängig von der FET-Ausgabe werden Grundlagen für die Stundenplanung in package "ttbase" vorbereitet.

## Neu: Druckausgabe

Stundenpläne können jetzt als PDF ausgegeben werden, aktuell die Klassentabellen und die Lehrertabellen. Dafür muss Typst installiert und als „typst“ aufrufbar sein. Zur Zeit wird die Typst-Datei als „data/resources/print_timetable.typ“ relativ zur Eingabedatei erwartet. Der Befehl wäre etwa:

```W365toFET -p path/to/sp001_w365.json```

Bei Erfolg wären die Ergebnisse dann im Ordner ```path/to/_pdf``` zu finden. Ein Fehlerbericht kann, wie oben, in der Log-Datei gefunden werden.
