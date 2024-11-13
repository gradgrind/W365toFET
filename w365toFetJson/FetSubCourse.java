package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Course;
import lombok.Getter;
import lombok.Setter;

import java.util.Objects;

@Getter
@Setter
public class FetSubCourse extends FetCourse {

  String superCourse;

  public FetSubCourse(Course course, String superCourseId) {
    super(course);
    setType("SubCourse");
    this.superCourse = superCourseId;
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (!(o instanceof FetSubCourse subCourse)) return false;
    if (!super.equals(o)) return false;
    return Objects.equals(superCourse, subCourse.superCourse);
  }

  @Override
  public int hashCode() {
    return Objects.hash(super.hashCode(), superCourse);
  }
}
