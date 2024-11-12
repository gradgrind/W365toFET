package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Subject;

public class FetSubject extends FetBaseObjectWithShortcutAndName {

  public FetSubject(Subject subject) {
    super(subject.getId(), "Subject", subject.getShortcut(), subject.getName());
  }
}
