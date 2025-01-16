# W365toFET

Build a FET file from the data supplied in a JSON file.

The name of the input file should ideally end with "_w365.json", such as "myfile_w365.json". This will enable a consistent automatic naming of the generated files.

**TODO:** For more flexibility, add the possibility of providing names for the generated files in the JSON input.

The files produced are saved in the same folder as the input file:

 - Log file: Contains error messages and warnings as well as information about the steps performed. Standard file name (given "myfile_w365.json" as input) is "myfile_w365.log".

 - FET file: The file to be fed to FET, standard name  (given "myfile_w365.json" as input) is "myfile.fet".

 - Map file: Correlates the FET Activity numbers to the Waldorf 365 Lesson references ("Id"). The standard name (given "myfile_w365.json" as input) is "myfile.map".

Note that, at present, the Activity and Room objects in the FET file have the corresponding Waldorf 365 references in their "Comments" fields.

First the input file is read and processed so that the data can be stored in a form independent of Waldorf 365. This form is managed in the "base" package, the primary data structure being the `DbTopLevel` `struct` defined in base/structures.go.

There are some useful pieces of information which are not stored directly in the basic data loaded from an input file, but which can be derived from it. The method `DbTopLevel.PrepareDb()` (in base/db_manager.go) performs the first of this processing and also checks for certain errors in the data.

For processing of timetable information there are further useful data structures which can be derived from the input data. This information is handled primarily in the "ttbase" package, its primary data structure being the `TtInfo` `struct` defined in ttbase/base.go.

A further stage of processing the timetable data is handled by the method `TtInfo.PrepareCoreData()`. This builds further data structures representing the allocation of resources, so that a number of errors in the data can be detected.

Finally, the data structures are used by the function `MakeFetFile` in the "fet" package to produce the XML-structure of the FET file and the reference mapping information to be stored in the map file.
