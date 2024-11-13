package de.ug_software.andromeda.io.fet;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class FetBaseObjectWithShortcutAndName extends FetBaseObjectWithShortcut {

  @JsonProperty(index=4)
  String Name;

  public FetBaseObjectWithShortcutAndName(String id, String type, String shortcut, String name) {
    super(id, type, shortcut);
    Name = name;
  }

}
