package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Absence;
import lombok.Data;

@Data
public class FetAbsence {

  int day;
  int hour;

  public FetAbsence(Absence absence) {
    this.day = absence.getDay();
    this.hour = absence.getHour();
  }
}
