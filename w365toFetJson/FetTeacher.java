package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Teacher;
import lombok.Getter;
import lombok.Setter;

import java.util.Comparator;
import java.util.List;

@Getter
@Setter
public class FetTeacher extends FetBaseObjectWithShortcutAndName {
  String firstname;
  int minLessonsPerDay;
  int maxLessonsPerDay;
  int maxDays;
  int maxGapsPerDay;
  int maxGapsPerWeek;
  int maxAfternoons;
  boolean lunchBreak = true;
  List<FetAbsence> absences;
  public FetTeacher(Teacher teacher) {
    super(teacher.getId(), "Teacher", teacher.getShortcut(), teacher.getName());
    firstname = teacher.getFirstname();
    minLessonsPerDay= teacher.getMinLessonsPerDay();
    maxLessonsPerDay = teacher.getMaxLessonsPerDay();
    maxDays = teacher.getMaxDays();
    maxGapsPerDay = teacher.getMaxWindowsPerDay();
    maxGapsPerWeek = teacher.getMaxGapsPerWeek();
    maxAfternoons = teacher.getNumberOfAfterNoonDays();
    absences = teacher.getAbsences().stream().map(FetAbsence::new)
            .sorted(Comparator.comparing(FetAbsence::getDay).thenComparing(FetAbsence::getHour))
            .toList();
  }
}
