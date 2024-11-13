package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Room;
import lombok.Getter;
import lombok.Setter;

import java.util.List;

@Getter
@Setter
public class FetRoomGroup extends FetBaseObjectWithShortcutAndName {

  List<String> rooms;
  public FetRoomGroup(Room room) {
    super(room.getId(), "Room", room.getShortcut(), room.getName());
    rooms = room.getRoomGroup().stream().map(Room::getId).toList();
  }
}
