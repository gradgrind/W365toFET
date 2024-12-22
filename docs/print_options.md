# Druckoptionen

> „Da es eine Vielzahl an Optionen geben sollte, fände ich es sinnvoller diese eventuell im w365TT Objekt (oder ein weiteres PrintOptions-Objekt)  in der JSOn-Datei anzugeben und die Konsolen-Parameter ggf. zu entfernen. Das wäre für mich der einfachere Weg.“

Ja, das wäre auf jeden Fall sinnvoll. Ich glaube, mir wäre ein neues PrintOptions-Objekt lieber als das w365TT-Objekt so zu erweitern.

## Druck von Raumplänen, Gesamtplänen nach Klassen, Lehrern und Räumen

Diese sind auf jeden Fall geplant.

## Angabe des Papierformats für Gesamtpläne.

Grundsätzlich sollte das kein Problem sein, da die Boxgrößen anhand der Papierdimensionen errechnet werden.

Ich hätte aber die Frage, ob dieser Parameter wirklich über die Optionen oder nicht lieber in den Typst-Skripten angegeben werden soll. Wenn die Seitengröße als Parameter angegeben wird, folgen weitere Fragen: wie es mit Seitenrändern, Schriftgrößen, usw. ist ... Das könnte dazu führen, dass es ziemlich viele neue Parameter geben müsste.

Es kann sein, dass es leichter ist, diese Parameter am Anfang der Typst-Skripte zu setzen. Das sind Werte, die wahrscheinlich nur einmal für eine Schule gesetzt werden, und oft wäre die Voreinstellung gut genug.

 - Generell scheint es in Typst einfacher mit den eigentlichen Dimensionen als mit den Namen von Papiergrößen zu arbeiten (also lieber 420x297 als „A3“). Typst kennt sehr viele dieser Namen, aber für die Berechnung der Boxgrößen sind die Maße notwendig. Wenn die Papiergröße als „A3“ angegeben wird, ist es nicht so leicht an die eigentlichen Größen zu kommen.
 
 - Die Versionen „Format A2 bestehend aus zwei A3-Seiten und A1 bestehend aus vier A3-Seiten“ wären wahrscheinlich besser durch dafür gemachte PDF-Programme oder -Bibliotheken (anhand der Einzelseiten) als durch Typst erstellt.

## Farben –  Hintergrundfarben von Lehrern, Klassen oder Fächern; Vordergrundfarbe dann für Kontrast.

 - Meine erste Reaktion: Ein Lesson könnte das neue Feld „background“ bekommen (Voreinstellung Weiß). Damit wäre dann auch der Vergleichsdruck abgedeckt, oder? Sie könnten alles in Waldorf 365 regeln.
 
 - Textfarbe Weiß oder Schwarz, je nach besserem Kontrast? Diese Wahl könnte automatisch erfolgen. Wenn ein komplizierteres Verfahren erwünscht ist, wäre mir ein weiteres neues Lesson-Feld „foreground“ lieber.

 - Problemfall: Es gibt mehrere Lehrer oder Klassen in einem Lesson. Wie wollen Sie damit umgehen? Im Prinzip könnte Typst wahrscheinlich über „gradients“ mehrere Farben in einer Box darstellen. Meine Erfahrung mit mehrfarbigen Boxen war aber nicht besonders erfreulich, da die Texte dann oft nur schwer lesbar sind – welche Kontrastfarbe sollte man wählen?

## Angabe des Raumes optional, Angabe der Gruppe optional

Es könnte in PrintOptions, die Felder „"NoRooms": true/false“ und „"NoGroups": true/false“ (Voreinstellung false) geben.

### weitere Gedanken dazu

Um wirklich flexibel zu sein, könnte man aber auch die Platzierung der verschiedenen Felder in einer Lesson-Box über Optionen regeln. Das wäre wieder eine Frage der Zuständigkeiten: Sollten diese Plätze in Waldorf 365, in W365toTypst oder in den Typst-Skripten entschieden werden?

Mein aktuelles Typst-Skript baut Boxen mit fünf möglichen Textfeldern: eins in der Mitte, die anderen in den vier Ecken. Wo die Fächer, Lehrer, usw. landen, wird in einer Go-Funktion entschieden, die für die konkrete Tabelle (Lehrer, Klasse, usw.) die Vorarbeit leistet. Eine Variante wäre, die Positionen der verschiedenen Felder als Optionen anzugeben, z.B.

```
"BOXTEXT_CLASS": {
 "C": "SUBJECT",   // Mitte
 "TL": "TEACHER",  // oben links
 "TR": "GROUP",    // oben rechts
 "BR": "ROOM"      // unten rechts
}
```

Wenn man Räume oder Gruppen nicht sehen wollte, könnte man deren Platzierungen einfach weglassen.

Für Gesamtpläne wären die Boxen wahrscheinlich kleiner und die Textpositionen möglicherweise anders, sodass andere Platzierungsoptionen nötig wären. Insgesamt könnten ziemlich viele neue Optionen zusammenkommen ...

Das sind auch Dinge, die vielleicht nur einmal für eine Schule entschieden werden, und vielleicht deswegen am besten in den Typst-Skripten gesetzt werden. Das heißt, es könnte im Typst-Skript diese Zuordnung der Felder zu den Textstellen geben.

## Angabe einer Schriftart

Fallbacks können benutzt werden.

Über die Typst-Kommandozeile (oder als Umgebungsvariable) kann ein Ordner mit Schriftarten (ttf, otf) angegeben werden.

Hier der Auszug aus der Dokumentation:

> When processing text, Typst tries all specified font families in order until it finds a font that has the necessary glyphs. ...

> The collection of available fonts differs by platform:

> ... Typst uses your installed system fonts or embedded fonts in the CLI, which are Libertinus Serif, New Computer Modern, New Computer Modern Math, and DejaVu Sans Mono. In addition, you can use the --font-path argument or TYPST_FONT_PATHS environment variable to add directories that should be scanned for fonts. The priority is: --font-paths > system fonts > embedded fonts. Run typst fonts to see the fonts that Typst has discovered on your system. Note that you can pass the --ignore-system-fonts parameter to the CLI to ensure Typst won't search for system fonts.

> Default: "libertinus serif"


## volle Lehrernamen (bzw. deren Abkürzung) anstelle von Kürzeln als Option

Das wäre natürlich irgendwie möglich, aber wie es dann sinnvoll gestaltet werden sollte, ist mir nicht klar. Wie kann ein Lehrername in der kleinen Box gut (lesbar) aussehen? Und wenn es mehrere Lehrer gibt?

## vollständigen Zeitraum oder nur Nummer der Stunde

Über die Option „WithTimes“ gibt es schon die Möglichkeit die Zeitangabe ein- bzw. auszuschalten. Über die Option „WithBreaks“ wird entschieden, ob nur die Unterrichtsstunden oder auch die Pausen dargestellt werden. Letztere ist für den Fall, dass es keine Zeitangaben für die Stunden gibt, notwendig. Diese Optionen könnten im PrintOptions-Objekt übergeben werden. Die Stundenbezeichnungen sind aktuell die Kürzel, die Tag-Bezeichnungen die Namen. Vielleicht sollte man in beiden Fällen Kürzel oder Namen wählen können? Oder die Bezeichnungen in Waldorf 365 festlegen und als Optionen übergeben?

```
 "days": ["Mo, "Di", ...]
 "hours": ["1", "2", ...]
```

Oder sowohl Kürzel als auch Namen an die Typst-Skripte übergeben und im Typst-Skript entscheiden?

Wie sehen Sie das?

## Ausgabe einer Legende unter dem Plan: Kürzel1=Name1, Kürzel2=Name2...

Im Prinzip wäre das möglich, die Idee überzeugt mich aber nicht besonders. Dadurch geht ggf. recht viel Platz verloren, den der Stundenplan u.U. gut gebrauchen könnte. Eine platzsparende Darstellung der Kürzel wäre vielleicht auch nicht besonders übersichtlich. Spricht etwas dagegen, stattdessen alle Kürzel geordnet und schön gestaltet in einem eigenen Dokument auszugeben?

## Ausgabe einer Überschrift, einer Unter-Überschrift, des Schulnamens

Das aktuelle Typst-Skript hat schon eine Box für Überschriften. Die Frage ist nur: Welche Felder sollten übergeben werden? Wo sie platziert werden und deren Formatierung sollten wahrscheinlich dem Typst-Skript überlassen werden. Mein Vorschlag für die Optionen wäre:

```
 "title": "Hauptüberschrift",
 "subtitle": "Entwurf Erstes Halbjahr | Letzte Änderung 15.06.2020 19:30 Uhr",
 "pageHeadingClass": "Klasse: %S",
 "pageHeadingTeacher": "%N (%S)",
 "pageHeadingRoom": "Raum: %N (%S)",
 "institution": "Freie Schule Mulmingen",
```

Im „pageHeadingXXX“ gäbe es dann über „%N“ und „%S“ die Möglichkeit Vollnamen und Kurznamen der jeweiligen Klasse, usw., einzubinden.

## Welche Pläne sollen gedruckt werden?

### eine Möglichkeit

Um einen Einzelplan (einer Klasse, eines Lehrers oder eines Raumes) auszudrucken, könnte es in PrintOptions folgenden Eintrag geben: „"printId": "Id des zu druckenden Objekts“

Wie soll die Ausgabedatei heißen? Sollte das Id enthalten sein, oder lieber etwas Generisches, z.B. "single"?

Wenn das printId-Feld leer ist, werden alle Pläne (Klassen, Lehrer, Räume und die Gesamtpläne gedruckt.

### eine Alternative

Für jeden möglichen Plan gibt es eine Option. Die Werte könnten true/false sein. Oder ein Dateiname könnte angegeben werden (leer => der Plan wird nicht gedruckt).

## Parameter direkt an Typst-Skripte übergeben

Bei einigen dieser Punkte gäbe es die Möglichkeit, dass sehr viele Paremeter entstehen, die mehr oder weniger direkt für die Typst-Skripte bestimmt wären. Diese zusätzlich in W365toTypst zu verarbeiten würde u.U. W365toTypst unnötig verkomplizieren. Falls diese Parameter tatsächlich entstehen sollten – also falls sie aus irgendeinem Grund wirklich in W365 und nicht direkt in den Typst-Skripten gesetzt werden sollten –, wäre es dann besser sie direkt an die entsprechenden Typst-Skripte zu übergeben. Das könnte als zusätzliche JSON-Datei-Eingabe oder eingebettet in der bestehenden Datei erfolgen.
