# W365toFET

```
W365toFET path/to/sp001_w365.json

    -> path/to/sp001_w365.log
    -> path/to/sp001.fet
    -> path/to/sp001.map
```

Die Stundenplan-Daten werden von der JSON-Datei eingelesen.

Ausgegeben wird eine FET-Datei im selben Ordner. Auch eine Logdatei (mit Fehlermeldungen, usw.) und eine Zuordnungsdatei für die FET-Activities werden erstellt.

## Aktueller Stand (02.12.2024)

Bis auf die „Constraint“-Elemente werden alle Elemente in `docs/stundenplanschnittstelle.md` in einigermaßen entsprechende FET-Strukturen übertragen.

Die weiteren Constraints werden jetzt anfänglich übersetzt.

In dieser Version haben Lehrer und Klassen den gleichen Ansatz für Mittagspausen: Eine der Mittagsstunden muss frei sein.

Es wurde bisher nur wenig getestet!

In dieser Version werden die Daten in eine etwas andere interne Struktur gebracht – W365-unabhängig – bevor sie übersetzt werden, siehe „base-package“. Auch unabhängig von der FET-Ausgabe werden Grundlagen für die Stundenplanung in package "ttbase" vorbereitet.
