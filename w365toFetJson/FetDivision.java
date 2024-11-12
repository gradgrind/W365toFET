package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.GradePartition;
import de.ug_software.andromain.model.basic.Group;
import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class FetDivision extends FetBaseObject {
  String name;
  List<String> groups;
  public FetDivision(GradePartition partition) {
    super(partition.getId(), "Division");
    name = partition.getName();
    groups = partition.getGroups().stream().map(Group::getId).toList();
  }
}
