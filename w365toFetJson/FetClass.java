package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.CalculationSchedule;
import de.ug_software.andromain.model.basic.Grade;
import lombok.Getter;
import lombok.Setter;

import java.util.Comparator;
import java.util.List;

@Getter
@Setter
public class FetClass extends FetBaseObjectWithShortcutAndName {
  int minLessonsPerDay;
  int maxLessonsPerDay;
  int maxGapsPerDay;
  int maxGapsPerWeek;
  int maxAfternoons;
  boolean lunchBreak = true;
  boolean forceFirstHour;
  List<FetAbsence> absences;
  int level;
  String letter;
  List<FetDivision> divisions;

  public FetClass(Grade grade, CalculationSchedule schedule) {
    super(grade.getId(), "Class", grade.getShortcut(), grade.getName());
    minLessonsPerDay = grade.getMinLessonsPerDay();
    maxLessonsPerDay = grade.getMaxLessonsPerDay();
    maxGapsPerDay = grade.getMaxWindowsPerDay();
    maxGapsPerWeek = grade.getMaxGapsPerWeek();
    maxAfternoons = grade.getNumberOfAfterNoonDays();
    absences = grade.getAbsences().stream().map(FetAbsence::new)
            .sorted(Comparator.comparing(FetAbsence::getDay).thenComparing(FetAbsence::getHour))
            .toList();
    divisions = grade.getGradePartitions().stream().filter(t ->
                    schedule.getLessons().stream().anyMatch(u -> u.getGroups().stream()
                            .anyMatch(g -> t.getGroups().contains(g))))
            .map(FetDivision::new).toList();
    forceFirstHour = grade.getForceFirstHour();
    letter = grade.getLetter();
    level = grade.getLevel();
  }
}
