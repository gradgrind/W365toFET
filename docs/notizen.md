# Notizen

### FirstAfternoonHour, MiddayBreak

Aktuell können diese sowohl von den „Hour“-Elementen als auch vom W365TT-Element eingelesen werden. Man sollte sich aber für eine Variante entscheiden.

### Min-Max-Eigenschaften der Lehrer und Klassen

Wenn diese Felder nicht vorhanden sind, wird der Wert -1 angenommen.

### LunchBreak (Lehrer und Klassen)

Wenn dieses Feld nicht vorhanden ist, wird „false“ angenommen.

### Subjects-Eigenschaft der Kurse (Course und SubCourse)

Aktuell wird programmintern mit einem **Subject**-Eigenschaft gearbeitet. Wenn aber die Eingabe die **Subjects**-Eigenschaft benutzt, wird dieses Feld eingelesen und notfalls umgewandelt.

Wenn es zwei oder mehr Fächer gibt, wird ein neues, künstliches Fach erstellt, das dann als Subject benutzt wird. Gleiche Fächerlisten werden dann durch das gleiche neue Fach ersetzt.

Es dürfen natürlich nicht beide Felder (Subject und Subjects) vorhanden sein.