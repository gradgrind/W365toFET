package de.ug_software.andromeda.io.fet;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.MapperFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import com.google.common.base.Joiner;
import de.ug_software.andromain.core.persistence.Crud;
import de.ug_software.andromain.model.TimedObject;
import de.ug_software.andromain.model.basic.*;
import lombok.Data;
import lombok.Getter;
import lombok.Setter;
import lombok.extern.slf4j.Slf4j;

import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Getter
@Setter
@Slf4j
public class W365ToFetJson {

  @JsonProperty(index = 0)
  FetSettings w365TT = new FetSettings();
  List<FetDay> days;
  List<FetHour> hours;
  List<FetTeacher> teachers;
  List<FetSubject> subjects;
  List<FetRoom> rooms;
  List<FetRoomGroup> roomGroups;
  List<FetClass> classes;
  List<FetGroup> groups;
  List<FetCourse> courses;
  List<FetSuperCourse> superCourses;
  List<FetEpochPlan> epochPlans;
  Set<FetSubCourse> subCourses;
  List<FetLesson> lessons;

  @Data
  private static class EpochplanTimeslot {
    int day;
    int hour;
    EpochPlan epochPlan;

    public EpochplanTimeslot(Lesson lesson) {
      this.day = lesson.getDay();
      this.hour = lesson.getHour();
      this.epochPlan = lesson.getEpochPlan();
    }
  }

  public W365ToFetJson(CalculationSchedule schedule) {
    // Only add objects actually needed for calculation
    days = Crud.read(Day.class).stream().map(FetDay::new).toList();
    hours = Crud.read(TimedObject.class).stream().map(FetHour::new).toList();
    teachers = schedule.getLessons().stream().flatMap(t -> t.getTeachers().stream())
            .distinct().map(FetTeacher::new).toList();
    subjects = schedule.getLessons().stream().flatMap(t -> t.getSubjects().stream())
            .distinct().map(FetSubject::new).toList();

    // add room groups that appear either as preferred rooms of a course
    // or as selected (local) room of al lesson:
    roomGroups = Stream.concat(schedule.getLessons().stream().flatMap(t -> t.getLocalRooms().stream()),
                    schedule.getLessons().stream().flatMap(t -> t.getCourse().getPreferredRooms().stream()))
            .filter(t -> !t.getRoomGroup().isEmpty()).distinct()
            .map(FetRoomGroup::new).toList();
    classes = schedule.getLessons().stream().flatMap(t -> t.getGroups().stream())
            .distinct().filter(t -> t instanceof Grade).map(Grade.class::cast)
            .map(t -> new FetClass(t, schedule)).toList();
    groups = schedule.getLessons().stream().flatMap(t -> t.getGroups().stream())
            .distinct().filter(t -> !(t instanceof Grade))
            .map(FetGroup::new).toList();

    // first add non epoch-courses:
    courses = schedule.getLessons().stream().flatMap(t -> Stream.of(t.getCourse()))
            .filter(t -> !(t instanceof EpochPlanCourse)).distinct()
            .map(FetCourse::new).toList();

    var epochPlanTimeslotToLessonsMap = mapEpochplanTimeslotsToLessons(schedule);

    // each super course corresponds to a group of epoch plan courses that do not collide.
    var distinctGroupsOfEpochPlanCourses = calculateDistinctGroupsOfEpochPlanCourses(epochPlanTimeslotToLessonsMap);
    superCourses = distinctGroupsOfEpochPlanCourses.stream().map(FetSuperCourse::new).toList();
    epochPlans = epochPlanTimeslotToLessonsMap.values().stream().flatMap(Collection::stream)
            .map(t -> ((EpochPlanCourse) t.getCourse()).getEpochenplan()).distinct()
            .map(FetEpochPlan::new).toList();

    // rooms are either preferred rooms or local rooms of normal lessons...
    rooms = Stream.concat(Stream.concat(schedule.getLessons().stream()
                                    .flatMap(t -> t.getCourse().getPreferredRooms().stream()),
                            schedule.getLessons().stream().flatMap(t -> t.getLocalRooms().stream())),
                    // ...or they are the set of all local rooms of all epochs
                    // in the corresponding epochplan of an epoch lesson
                    distinctGroupsOfEpochPlanCourses.stream().flatMap(Collection::stream).flatMap(t -> t.getEpochenplan().getLessons().stream())
                            .flatMap(u -> u.getLocalRooms().stream()))
            .distinct().filter(t -> t.getRoomGroup().isEmpty())
            .map(FetRoom::new).toList();

    subCourses = new HashSet<>();
    for (var courseSet : distinctGroupsOfEpochPlanCourses) {
      for (var course : courseSet)
        subCourses.addAll(course.getEpochenplan().getLessons().stream()
                .filter(t -> t.getGroups().stream()
                        .anyMatch(g -> Objects.equals(course.getGrade(), g.getGrade())))
                .flatMap(t -> Stream.of(t.getCourse()))
                .map(u -> new FetSubCourse(u, calculateSuperCourseId(courseSet))).toList());
    }
    lessons = Stream.concat(
                    // lessons are either normal lessons...
                    schedule.getLessons().stream().filter(t -> !(t.getCourse() instanceof EpochPlanCourse))
                            .map(t -> new FetLesson(t, t.getCourse().getId())),

                    // ...or epochplan lessons:
                    epochPlanTimeslotToLessonsMap.values().stream().map(lessonList -> new FetLesson(lessonList.get(0),
                            calculateSuperCourseId(lessonList.stream().map(Lesson::getCourse)
                                    .map(EpochPlanCourse.class::cast).toList()))))
            .collect(Collectors.toList());
  }

  private static String calculateSuperCourseId(Collection<EpochPlanCourse> courseSet) {
    return Joiner.on(",")
            .join(courseSet.stream().map(EpochPlanCourse::getFetId)
                    .toList().stream().sorted().toList());
  }

  private static HashSet<Set<EpochPlanCourse>> calculateDistinctGroupsOfEpochPlanCourses(HashMap<EpochplanTimeslot, List<Lesson>> epochPlanTimeslotToLessonsMap) {
    var distinctGroupsOfEpochPlanCourses = new HashSet<Set<EpochPlanCourse>>();
    for (var entry : epochPlanTimeslotToLessonsMap.entrySet())
      distinctGroupsOfEpochPlanCourses.add(entry.getValue().stream().map(Lesson::getCourse)
              .map(EpochPlanCourse.class::cast).collect(Collectors.toSet()));
    return distinctGroupsOfEpochPlanCourses;
  }

  private static HashMap<EpochplanTimeslot, List<Lesson>> mapEpochplanTimeslotsToLessons(CalculationSchedule schedule) {
    var epochPlanCourseLessons = schedule.getLessons().stream()
            .filter(t -> t.getCourse() instanceof EpochPlanCourse).distinct().toList();
    var epochPlanTimeslotToLessonsMap = new HashMap<EpochplanTimeslot, List<Lesson>>();
    for (var lesson : epochPlanCourseLessons)
      epochPlanTimeslotToLessonsMap.computeIfAbsent(new EpochplanTimeslot(lesson), t -> new ArrayList<>()).add(lesson);
    return epochPlanTimeslotToLessonsMap;
  }

  public String toJson() throws JsonProcessingException {
    ObjectMapper mapper = new ObjectMapper()
            .configure(MapperFeature.SORT_PROPERTIES_ALPHABETICALLY, true)
            .enable(SerializationFeature.INDENT_OUTPUT);
    mapper.registerModule(new JavaTimeModule()).disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    return mapper.writeValueAsString(this);
  }

}
