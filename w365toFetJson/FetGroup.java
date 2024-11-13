package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Group;

public class FetGroup extends FetBaseObjectWithShortcut {

  public FetGroup(Group group) {
    super(group.getId(), "Group", group.getShortcut());
  }
}
