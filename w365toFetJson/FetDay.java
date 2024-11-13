package de.ug_software.andromeda.io.fet;

public class FetDay extends FetBaseObjectWithShortcutAndName {

  public FetDay(de.ug_software.andromain.model.basic.Day day) {
    super(day.getId(), "Day", day.getShortcut(), day.getName());
  }
}
