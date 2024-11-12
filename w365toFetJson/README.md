# W365toJson

The json string is produced as follows:
```
var config = new PlanningConfiguration();
config.setFixExistingLessons(PlanningConfigurationDescription.FIX_ALL_EXISTING_AND_LEAVE_FIXED);
var calculationSchedule = new CalculationSchedule(startingScheduleWithFixedLessons, config);
var jsonString = new W365ToFetJson(calculationSchedule).toJson();
``` 
