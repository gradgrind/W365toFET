package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.TimedObject;
import lombok.Getter;
import lombok.Setter;

import java.time.LocalTime;

@Getter
@Setter
public class FetHour extends FetBaseObjectWithShortcutAndName {
  LocalTime start;
  LocalTime end;
  public FetHour(TimedObject timedObject) {
    super(timedObject.getId(), "Hour", timedObject.getShortcut(), timedObject.getName());
    start = timedObject.getStart();
    end = timedObject.getEnd();
  }
}
