package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.DomainRoot;
import de.ug_software.andromeda.calc.schedule.constraints.ScheduleFixConstraintsProvider;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
public class FetSettings {

  int firstAfternoonHour = ScheduleFixConstraintsProvider.getFirstAfternoonHour();
  List<Integer> middayBreak = new ArrayList<>();

  public FetSettings(){
    for(var hour=0;hour<DomainRoot.giveNumberOfLessonsPerDay();hour++){
      if(DomainRoot.lessonPeriods().get(hour).isMiddayBreak())
        middayBreak.add(hour);
    }
  }
}
