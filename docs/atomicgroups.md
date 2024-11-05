# "Atomic" Groups

For automatic timetabling algorithms it is often helpful to have special, potentially abstract, student groups which may be placed in the plan independently of any other such group. Of course, on a physical level that could correspond to a single student. However, this approach has the disadvantage of coupling the timetable to the number of students, whereas we will normally be dealing with only classes and, quite possibly, teaching groups within classes – relatively independent of actual students.

The difficulty arises when a class may be divided in more than one way. Let's say a class has one division (A, B) in Maths and English. Then there is another division (P, Q) for Art and Music. A lesson belonging to one of these divisions may not be placed parallel to (in the same time slot as) a lesson belonging to another division.For example, Maths-A could be placed parallel to English-B, but not Art-Q.

One way of dealing with this would be to check the division of each lesson's group when placing it, comparing it to the division of a lesson already in the time slot.

Another approach, which may have advantages from an algorithmic point of view, is to check for clashes between individual groups only, in effect disregarding the divisions. This something of a bottom-up approach vs. the more top-down approach of the division-checking.

For this view of things, it is useful to have the class divided into subgroups which can be placed independently of each other. In the example above, there would be four of these: AP, AQ, BP and BQ.

To generate these subgroups – which I here call "atomic" (indivisible) groups – the Cartesian product of the groups within the divisions is generated. Let's take a more complicated example, with divisions with d1 = (A, B), d2 = (P, Q, R) and d3 = (X, Y). The Cartesian product, d1 x d2 x d3, is then:

```
    A-P-X
    A-P-Y
    A-Q-X
    A-Q-Y
    A-R-X
    A-R-Y
    B-P-X
    B-P-Y
    B-Q-X
    B-Q-Y
    B-R-X
    B-R-Y
```
There are 2 x 3 x 2 = 12 atomic groups. These can be generated using nested for-loops:

```
aglist = []         // empty list of (empty list of Groups)
aglist.append([])   // add (empty list of Groups) -> [[]]
FOR division IN divisions:
    // Build a new structure
    aglistx = []    // empty list of (empty list of Groups)
    // Extend each entry in aglist for each of the div-Groups
    FOR ag IN aglist:
        FOR group in division:
            agx = ag.copy()
            agx.append(group)
            aglistx.append(agx)
        ENDFOR
    // Replace original list and go to next division
    aglist = aglistx
    ENDFOR
ENDFOR

// aglist = [[A,P,X], [A,P,Y], [A,Q,X], [A,Q,Y], [A,R,X], [A,R,Y],
//           [B,P,X], [B,P,Y], [B,Q,X], [B,Q,Y], [B,R,X], [B,R,Y]]
```

Each of the "real" groups (A, B, P, Q, R, X, Y) corresponds to a set of these atomic groups:

```
    A = {A-P-X, A-P-Y, A-Q-X, A-Q-Y, A-R-X, A-R-Y}
    B = {B-P-X, B-P-Y, B-Q-X, B-Q-Y, B-R-X, B-R-Y}
    P = {A-P-X, A-P-Y, B-P-X, B-P-Y}
    Q = {A-Q-X, A-Q-Y, B-Q-X, B-Q-Y}
    R = {A-R-X, A-R-Y, B-R-X, B-R-Y}
    X = {A-P-X, A-Q-X, A-R-X, B-P-X, B-Q-X, B-R-X}
    Y = {A-P-Y, A-Q-Y, A-R-Y, B-P-Y, B-Q-Y, B-R-Y}
```

This mapping can be built by searching the atomic groups for each of the real groups, but there is also an approach which takes advantage of the order in which they are generated:

 - In the last column of the atomic groups, the groups (from the last – the third – division) each appear once before switching to the next.
 - In the next-to-last column, the groups (now from the second division) appear `2 = 2 x 1` times before switching to the next (two being the number of groups in the third division).
 - In the first column, the groups (from the first division) appear `6 = 3 x 2 x 1` times before switching to the next (there being three groups in the second division).

 A corresponding algorithm might then look something like this:
 
```
// Note: list indexing starts at 0, not 1.

group_ag_map = map(Group, [])   // group -> list of atomic-group-indexes
FOR g in real_groups:           // initialize lists for each real group
    group_ag_map[g] = []
ENDFOR
count = 1   // number of repetitions of each group
div_index = length(divisions)
WHILE div_index > 0:
    div_index = div_index - 1
    div_groups = divisions[div_index]
    agi = 0     // atomic-group-index
    WHILE agi < length(aglist):
        FOR group IN div_groups:
            REPEAT count times:
                group_ag_map[group].append(agi)
                agi = agi + 1
            ENDREPEAT
        ENDFOR
    ENDWHILE
    count = count * length(div_groups)
ENDWHILE

// group_ag_map[A] = [0, 1, 2, 3, 4, 5]
// group_ag_map[B] = [6, 7, 8, 9, 10, 11]
// group_ag_map[P] = [0, 1, 6, 7]
// group_ag_map[Q] = [2, 3, 8, 9]
// group_ag_map[R] = [4, 5, 10, 11]
// group_ag_map[X] = [0, 2, 4, 6, 8, 10]
// group_ag_map[Y] = [1, 3, 5, 7, 9, 11]
``` 

By appending the atomic groups themselves rather than their indexes, the result would map directly to the atomic groups:

``` 
// group_ag_map[group].append(agi)
// ->
group_ag_map[group].append(aglist[agi])
``` 
