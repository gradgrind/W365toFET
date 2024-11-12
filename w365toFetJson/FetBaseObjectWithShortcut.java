package de.ug_software.andromeda.io.fet;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class FetBaseObjectWithShortcut extends FetBaseObject {

  @JsonProperty(index=3)
  String Shortcut;
  public FetBaseObjectWithShortcut(String id, String type, String shortcut) {
    super(id, type);
    Shortcut = shortcut;
  }
}
