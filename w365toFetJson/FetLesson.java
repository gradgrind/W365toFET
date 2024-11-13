package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Lesson;
import de.ug_software.andromain.model.basic.Room;
import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class FetLesson extends FetBaseObject {
  String course;
  int duration;
  int day;
  int hour;
  boolean fixed;
  List<String> localRooms;

  public FetLesson(Lesson lesson, String courseId) {
    super(lesson.getId(), "Lesson");
    course = courseId;
    duration = lesson.isWillBeDoubleLesson()?2:1;
    day = lesson.getDay();
    hour = lesson.getHour();
    fixed = lesson.getFixed();
    localRooms = lesson.getLocalRooms().stream().map(Room::getId).toList();
  }
}
