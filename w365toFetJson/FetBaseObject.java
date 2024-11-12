package de.ug_software.andromeda.io.fet;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class FetBaseObject {
  @JsonProperty(index=1)
  String id;

  @JsonProperty(index=2)
  String type;

  public FetBaseObject(String id, String type){
    this.id = id;
    this.type = type;
  }
}
