package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.EpochPlan;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class FetEpochPlan extends FetBaseObjectWithShortcutAndName {

  public FetEpochPlan(EpochPlan epochPlan) {
    super(epochPlan.getId(), "EpochPlan", epochPlan.getShortcut(), epochPlan.getName());
  }
}
