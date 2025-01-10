/* This is a script to generate a view of the complete timetable for all
 * classes, teachers or rooms in a fairly compact form. Each of the items
 * has a row showing its weekly timetable.
 *
 * The basic idea is to use a Typst-table to manages the headers, lines and
 * background colouring. The tiles of the timetable activities are overlaid
 * on this using the table cell boundary coordinates for orientation.
 *
 * A space (block) is left at the top of each page for a page title. This
 * is in the space left free by the TITLE_HEIGHT value.
 * The rest of the page will be used for the table, adjusting the cell size
 * to fit.
 *
 * It is quite likely that there will be too many items for a single page. In
 * this case, the item list is divided – the available space on the page is
 * known, as are the row heights, so this division is fairly straightforward.
 */

// To use a different font:
// CHANGE_UG
#set text(font: ("Nunito","DejaVu Sans"))
// If the font is not installed on the system, the .ttf or .otf files can be
// placed in "typst_files/_fonts".

// For A3 paper
#let PAGE_HEIGHT = 297mm
#let PAGE_WIDTH = 420mm
#let PAGE_BORDER = (top:15mm, bottom: 15mm, left: 15mm, right: 15mm)
#let TITLE_HEIGHT = 15mm
#let H_HEADER_HEIGHT1 = 10mm
#let H_HEADER_HEIGHT2 = 10mm

#let ROW_HEIGHT = 12mm
#let ROW_HEIGHT_CLASS = 30mm // larger because of divisions
#let DAY_SPACING = 2mm

#let HEADING_TO_PLAN_DISTANCE = 4mm
#let PLAN_TO_FOOTER_DISTANCE = 3mm

#let CELL_BORDER = 0.5pt
#let BIG_SIZE = 24pt
#let NORMAL_SIZE = 13pt
#let CELL_TEXT_SIZE = 10pt
#let DAY_SIZE = 13pt
#let HOUR_SIZE = 9pt

#let FRAME_COLOUR = "#707070"
#let HEADER_COLOUR = "#ffffff"
#let EMPTY_COLOUR = "#ffffff"

#let JOINSTR = ","
#let CLASS_GROUP_JOIN = "."

// Field placement fallbacks
#let boxText = (
    Class: (
        M: "SUBJECT",
        T: "TEACHER",
        B: "GROUP",
        //B: "ROOM",
    ),
    Teacher: (
        M: "GROUP",
        T: "SUBJECT",
        B: "TEACHER",
        //B: "ROOM",
    ),
    Room: (
        M: "GROUP",
        T: "SUBJECT",
        B: "TEACHER",
    ),
)

// Document title fallbacks
#let titleFallbacks = (
    Class: "Gesamtstundenplan der Klassen",
    Teacher: "Gesamtstundenplan der Lehrkräfte",
    Room: "Gesamtstundenplan der Räume",
)

// Row heading fallbacks
// Folgende Zeichen werden ersetzt:
// %0 = Kürzel des Lehrer / Raums ...
// %1 = Nachname des Lehrers, Name der Klasse oder des Raums
// %2 = Vorname des Lehrers
#let rowHeadings = (
    Class: "%0",
    Teacher: "%2 %1",
    Room: "%1",
)

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

#let emptyColour = rgb(EMPTY_COLOUR)
#let headerColour = rgb(HEADER_COLOUR)

#let PLAN_AREA_HEIGHT = (PAGE_HEIGHT - PAGE_BORDER.top
    - PAGE_BORDER.bottom - TITLE_HEIGHT - HEADING_TO_PLAN_DISTANCE - PLAN_TO_FOOTER_DISTANCE)
#let PLAN_AREA_WIDTH = (PAGE_WIDTH - PAGE_BORDER.left
    - PAGE_BORDER.right)

#let H_HEADER_HEIGHT = H_HEADER_HEIGHT1 + H_HEADER_HEIGHT2
#let TABLE_BODY_HEIGHT = PLAN_AREA_HEIGHT - H_HEADER_HEIGHT

#let xdata = json(sys.inputs.ifile)
#let typstMap = xdata.at("Typst", default: (:))

#let DAYS = ()
#for ddata in xdata.Info.Days {
    //TODO: Which field to use
    DAYS.push(ddata.Name)
}

#let HOURS = ()
#for hdata in xdata.Info.Hours {
    HOURS.push(hdata.Short)
}

#let classList = xdata.Info.ClassNames
#let teacherList = xdata.Info.TeacherNames
#let roomList = xdata.Info.RoomNames
#let subjectList = xdata.Info.SubjectNames

#let listListToMap(listList) = {
    let map = (:)
    for list in listList {
        map.insert(list.at(0), list)
    }
    map
}

#let legendItems(ilist) = {
    let items = ()
    for item in ilist {
        if item.len() == 3 {
            // Assume teacher
            items.push(box(item.at(0)+" = "+item.at(2)+" "+item.at(1)))
        } else {
            items.push(box(item.at(0)+" = "+item.at(1)))
        }
    }
    items.join(", ")
}

#let nameMap = (
    "Class": listListToMap(classList),
    "Teacher": listListToMap(teacherList),
    "Room": listListToMap(roomList),
//    "Subject": listListToMap(subjectList),
)

// Type of table ("Class", "Teacher" or "Room")
#let tableType = xdata.TableType
// Row headings
#let rhead = typstMap.at("RowHeading", default: "")
#if rhead == "" {
    rhead = rowHeadings.at(tableType)
}
// Determine the field placements in the tiles
#let fieldPlacements = typstMap.at("FieldPlacement", default: (:))
#if fieldPlacements.len() == 0 {
    // fallback
    fieldPlacements = boxText.at(tableType, default: (:))
}

// +++ Set up the table
#let ndays = DAYS.len()
#let nhours = HOURS.len()
#let pcols = ndays * nhours
#let daySpacing = DAY_SPACING * (ndays - 1)


#show table.cell: it => {
    if it.y == 0 {
        set text(size: DAY_SIZE, weight: "bold")
        align(center + horizon, it.body.at("text", default: ""))
    } else if it.y == 1 {
        set text(size: HOUR_SIZE, weight: "bold")
        align(center + horizon, it.body.at("text", default: ""))
    } else if it.x == 0 {
        set text(size: NORMAL_SIZE, weight: "bold")
        align(center + horizon, it.body.at("text", default: ""))
    } else {
        it
    }
}

#let shrinkwrap(
    width,
    textc,
    tsize: NORMAL_SIZE,
    bold: false,
    hpos: center,
    vpos: horizon,
) = {
    let wt = "regular"
    if bold { wt = "bold" }
    context {
        let t
        let text2 = textc
        let xwhile = true
        while true {
            t = text(size: tsize, weight: wt, text2)
            let s = measure(t)
            if s.width > width * 0.9 {
                let scl = (width * 0.9 / s.width)
                // Shorten if the text is too long
                if xwhile and scl < 0.5 {
                    // Take just the first items in the list and add a "+".
                    xwhile = false
                    let tlist = textc.split(JOINSTR)
                    let n = int(tlist.len() * 2 * scl)
                    if n == 0 {
                        // Long first entry, truncate it
                        tlist = tlist.at(0).clusters()
                        n = int(tlist.len() * 2 * scl)
                        text2 = tlist.slice(0, n).join("") + "\u{00A0}+"
                    } else {
                        text2 = tlist.slice(0, n).join(JOINSTR) + "\u{00A0}+"
                    }
                    continue
                }
                t = text(size: scl * tsize, weight: wt, text2)
            }
            break
        }
        place(vpos + hpos, t)
    }
}

#if tableType == "Class" {
    ROW_HEIGHT = ROW_HEIGHT_CLASS
}

// Get the vertical headers.
#let vheaders = ()
#for page in xdata.Pages {
    let nameList = nameMap.at(tableType).at(page.Short)
    let rh = rhead.replace(regex("%([0-2])"),
        m => nameList.at(int(m.captures.first())))
    vheaders.push(rh)
}

#context {

    let V_HEADER_WIDTH = vheaders.fold(0mm, (max, it) => {
    let w = measure(text(size:NORMAL_SIZE, weight:"bold")[#it]).width
    if w > max { w } else { max }
    }) + 3mm

    let colwidth = (PLAN_AREA_WIDTH - V_HEADER_WIDTH - daySpacing) / pcols

    let cell_inset = 0
    let cell_height = ROW_HEIGHT


    // Prepare vertical header and row sizes and boundaries
    let rowsPerPage = int(TABLE_BODY_HEIGHT / ROW_HEIGHT)

    let hlines = (H_HEADER_HEIGHT,)
    let y = H_HEADER_HEIGHT
    let i = 0
    while i < rowsPerPage {
        i += 1
        y += ROW_HEIGHT
        hlines.push(y)
    }

    show heading: it => text(weight: "bold", size: BIG_SIZE,
        bottom-edge: "descender",
        pad(left: 0mm, it))

    // Determine the document title
    let doctitle = typstMap.at("Title", default: "")
    if doctitle == "" {
        // fallback
        doctitle = titleFallbacks.at(tableType, default: "")
    }
    let subtitle = typstMap.at("Subtitle", default: "")
    let lastChange = typstMap.at("LastChange", default: "")
    if subtitle == "" {
        subtitle = lastChange
    } else if lastChange != "" {
        subtitle += " · " + lastChange
    }

    set page(height: PAGE_HEIGHT, width: PAGE_WIDTH,
        //numbering: "1/1",
        margin: PAGE_BORDER,
    )


    // Prepare horizontal header, also column sizes and boundaries
    let dheader = ([],)
    let pheader = ([],)
    let vlines = (V_HEADER_WIDTH,)
    let x = V_HEADER_WIDTH
    let tcolumns = (V_HEADER_WIDTH,)
    let hcellfill = ()
    let firstDay = true
    for d in DAYS {
        if firstDay {
            firstDay = false
        } else {
            dheader.push(table.cell(rowspan: 2, ""))
            x += DAY_SPACING
            vlines.push(x)
            tcolumns.push(DAY_SPACING)
            hcellfill.push(table.cell(fill: headerColour, ""))
        }
        dheader.push(table.cell(colspan: nhours, d))
        for p in HOURS {
            pheader.push(p)
            x += colwidth
            vlines.push(x)
            tcolumns.push(colwidth)
            hcellfill.push([])
        }
    }

    let ttvcell(
        row,
        irow,
        Day: 0,
        Hour: 0,
        Duration: 1,
        Offset: 0,
        Fraction: 1,
        Total: 1,
        Subject: "",
        Groups: (),
        Teachers: (),
        Rooms: (),
        Background: "",
        Footnote: "",
    ) = {
        // Determine grid lines
        let ix = Day * nhours + Hour
        ix += Day
        let x0 = vlines.at(ix) - CELL_BORDER
        let y0 = hlines.at(row) - CELL_BORDER
        // Prepare texts
        let clist = () // Class part only
        let glist = ()
        let cglist = ()
        for (c, g) in Groups {
            clist.push(c)
            let cg = c
            if g != "" {
                cg += CLASS_GROUP_JOIN + g
            }
            if xdata.Pages.at(irow).Short == c {
                if g != "" {
                    glist.push(g)
                }
            } else {
                glist.push(cg)
            }
            cglist.push(cg)
        }
        clist = clist.dedup() // Remove duplicates
        let texts = (
            SUBJECT: Subject,
            CLASS: clist.join(JOINSTR),
            TEACHER: Teachers.join(JOINSTR),
            ROOM: Rooms.join(JOINSTR),
            GROUP: glist.join(JOINSTR),
            CLASS_WITH_GROUP: cglist.join(JOINSTR),
        )
        
        let ctext = texts.at(fieldPlacements.at("M", default: ""), default: "") 
        let ttext = texts.at(fieldPlacements.at("T", default: ""), default: "") + Footnote
        let btext = texts.at(fieldPlacements.at("B", default: ""), default: "")

        if Background == "" {
            Background = "#FFFFFF"
        }
        let bg = rgb(Background)
        // Get text colour
        // 1) This converts the background to grey-scale and uses a threshold:
        let bw = luma(bg)
        let textcolour = if bw.components().at(0) < 75% { white } else { black }
        set text(textcolour)

        // Determine size and offset of tile
        let w = colwidth * Duration
        let hfrac = cell_height * Fraction / Total
        let yshift = cell_height * Offset / Total
        // Shrink excessively large components.
        let b = box(
            fill: bg,
            stroke: (paint: rgb(FRAME_COLOUR),thickness:CELL_BORDER),
            inset: 2pt,
            height: hfrac,
            width: w,
        )[
            #shrinkwrap(w, ttext, tsize: CELL_TEXT_SIZE, hpos: left, vpos: top)
            #shrinkwrap(w, ctext, tsize: CELL_TEXT_SIZE, bold: true)
            #shrinkwrap(w, btext, tsize: CELL_TEXT_SIZE, hpos: right, vpos: bottom)
        ]
        place(top + left,
            dx: x0 + CELL_BORDER,
            dy: y0 + CELL_BORDER + yshift,
            b
        )
    }


    // +++ Divide the data into pages
    let irow = 0
    let rows = xdata.Pages
    let nrows = rows.len()
    let xrows = ()
    let iy = 0
    let aix = 0
    // Count page numbers
    let pageno = 1
    let pagetotal = int((rows.len() + rowsPerPage - 1) / rowsPerPage)
    let legend = typstMap.at("Legend", default: (:))
    if legend.len() != 0 {
        pagetotal += 1
    }

    let headerBlock(pagenum) = block(height: TITLE_HEIGHT, above: 0mm, below: 4mm, inset: 0mm)[
        #place(top)[= #doctitle]
        #place(top)[#h(1fr) #text(17pt, xdata.Info.Institution)]
        #place(bottom)[
            #subtitle
            #h(1fr)
            #pagenum / #pagetotal
        ]
    ]

    while irow < nrows {

        let trows = ((H_HEADER_HEIGHT1, H_HEADER_HEIGHT2) + (ROW_HEIGHT,))

        //let row = rows.at(irow)
        let rh = vheaders.at(irow)
        irow += 1

        xrows += (rh,) + hcellfill
        iy += 1
        if iy == rowsPerPage or irow == nrows{
            // Page done

            headerBlock(pageno)

            box([
                #table(
                    columns: tcolumns,
                    rows: trows,
                    gutter: 0pt,
                    stroke: (paint: rgb(FRAME_COLOUR), thickness: CELL_BORDER) ,
                    inset: 0pt,
                    fill: (x, y) =>
                        if y > 1 and x > 0 {
                            emptyColour
                        } else {
                            headerColour
                        },
                    table.header(
                        ..dheader, ..pheader,
                    ),
                    ..xrows,
                )

                #let rix = 0
                #while aix < irow {
                    for a in rows.at(aix).Activities {
                        ttvcell(rix, aix, ..a)
                    }
                    rix += 1
                    aix += 1
                }
            ])

            if irow != nrows {
                pagebreak()
            }
            xrows = ()
            iy = 0
            pageno += 1
        }
    }

    // Deal with the legend, if any
    if legend.len() != 0 {
        pagebreak()
        
        headerBlock(pageno)
        
        let legendRemark = legend.at("Remark", default: "")
        let footnotesLegend = legend.at("Footnotes", default: ())
        let subjectsLegend = legend.at("Subjects", default: false)
        let teachersLegend = legend.at("Teachers", default: false)
        let roomsLegend = legend.at("Rooms", default: false)

        text(10pt)[
            Hallo Welt!
            
            #if legendRemark != "" [*Hinweis:* #legendRemark \ ]
            #if footnotesLegend.len() != 0 [*Anmerkungen:* #legendItems(footnotesLegend) \ ]
            #if subjectsLegend [*Fächer:* #legendItems(subjectList) \ ]
            #if teachersLegend [*Lehrkräfte:* #legendItems(teacherList) \ ]
            #if roomsLegend [*Räume:* #legendItems(roomList) ]
        ]
    
    }
}
