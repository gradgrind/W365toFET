# Reading the JSON file from Waldorf 365

The structure of the input file (reflected in the structures in "w365tt/structures.go") is quite close to that of the structures in "base/structures.go", but there are some differences. The functions in package "w365tt" perform the operations necessary to read the data in the input file into these internal structures.

The function `ReadJSON` in "w365tt/fromjson.go" reads the input file into the structures in "w365tt/structures.go". It is called from `LoadJSON`, which then performs the processing necessary to transfer the data into the structures in "base/structures.go". `LoadJSON` is called from "cmd/W365toFET/main.go".

Keeping the `struct`s in "w365tt/structures.go" separate from the internal ones in "base/structures.go" allows some flexibility in the naming and contents of the input data. By adding tags like ``` `json:"Shortcut"` ``` to a field descriptor, the JSON field name can be changed without changing any internal data or code other than this tag.

Some of the structures need little or no translation, so they can be more or less directly copied. There is a suite of methods on `base.DbTopLevel` whose names start with "New" (`NewDay`, `NewHour`, etc.), defined in "base/db_manager.go", which call the method `addElement` to add an entry to the `Elements` mapping. This allows easy access to the element's internal `struct` via the element's `Id` (the `Id` fields from the Waldorf 365 input are retained for this purpose). If new elements (not in the input data) are created, new `Id`s are created for them.

## Blocked time-slots and maximum-afternoons constraints

When reading the data for elements with `MaxAfternoons` constraints (`Class` and `Teacher`) the `handleZeroAfternoons` method is called to block all afternoons (via `NotAvailable`) when the value is 0, the `MaxAfternoons` then being set in that case to -1 (constraint disabled). This method is also used to convert/build the `NotAvailable` structure /list of `TimeSlot`s) for the rooms, even though these have no `MaxAfternoons` constraint.

## Reading Class data

When reading the classes from the input file, some checks on the groups and class-divisions are made, to ensure that they meet with expectations. Also a new `Group` element is created for the whole class. This is significant for the courses which concern whole classes. In Waldorf 365 these would reference a `Class` element rather than a `Group` element. The use of such different types as target can be a bit troublesome, so these are converted to reference the newly created `Group` elements for whole classes.

### `GroupRefMap`

To simplify the handling of the class/group references in the courses of the Waldorf 365 data, this mapping (a field of the `w365tt.DbTopLevel` `struct`) is constructed. It maps Waldorf 365 `Group` references to themselves and Waldorf 365 `Class` references to references to the newly created `Group` elements.
