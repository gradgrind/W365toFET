package de.ug_software.andromeda.io.fet;

import com.google.common.base.Joiner;
import de.ug_software.andromain.model.basic.EpochPlan;
import de.ug_software.andromain.model.basic.EpochPlanCourse;
import lombok.Getter;
import lombok.Setter;

import java.util.Set;

@Setter
@Getter
public class FetSuperCourse extends FetBaseObject {

  private String epochPlan;

  public FetSuperCourse(Set<EpochPlanCourse> courses) {
    super(Joiner.on(",").join(courses.stream()
                    .map(EpochPlanCourse::getFetId).toList().stream().sorted().toList())
            , "SuperCourse");
    epochPlan = Joiner.on(",")
            .join(courses.stream().map(EpochPlanCourse::getEpochenplan).distinct()
                    .map(EpochPlan::getId).toList());
  }
}
