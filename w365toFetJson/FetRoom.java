package de.ug_software.andromeda.io.fet;

import de.ug_software.andromain.model.basic.Room;
import lombok.Getter;
import lombok.Setter;

import java.util.Comparator;
import java.util.List;

@Getter
@Setter
public class FetRoom extends FetBaseObjectWithShortcutAndName {

  List<FetAbsence> absences;
  public FetRoom(Room room) {
    super(room.getId(), "Room", room.getShortcut(), room.getName());
    absences = room.getAbsences().stream().map(FetAbsence::new)
            .sorted(Comparator.comparing(FetAbsence::getDay).thenComparing(FetAbsence::getHour))
            .toList();
  }
}
