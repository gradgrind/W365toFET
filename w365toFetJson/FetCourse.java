package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.*;
import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class FetCourse extends FetBaseObject {
  List<String> subjects;
  List<String> groups;
  List<String> teachers;
  List<String> preferredRooms;

  public FetCourse(Course course) {
    super(course.getId(), "Course");
    subjects = course.getSubjects().stream().map(Subject::getId).distinct().toList();
    groups = course.getGroups().stream().map(Group::getId).distinct().toList();
    teachers = course.getTeachers().stream().map(Teacher::getId).distinct().toList();
    preferredRooms = course.getPreferredRooms().stream().map(Room::getId).distinct().toList();
  }
}
