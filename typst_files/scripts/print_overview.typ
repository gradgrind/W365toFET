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
#let V_HEADER_WIDTH = 30mm
#let V_HEADER_WIDTH_CLASS = 15mm // smaller for classes
#let ROW_HEIGHT = 12mm
#let ROW_HEIGHT_CLASS = 30mm // larger because of divisions

#let CELL_BORDER = 0.5pt
#let BIG_SIZE = 24pt
#let NORMAL_SIZE = 13pt
#let CELL_TEXT_SIZE = 10pt
#let DAY_SIZE = 13pt
#let HOUR_SIZE = 9pt

#let FRAME_COLOUR = "#707070"
#let HEADER_COLOUR = "#f0f0f0"
#let EMPTY_COLOUR = "#ffffff"

#let JOINSTR = ","

// Field placement fallbacks
#let boxText = (
    Class: (
        m: "SUBJECT",
        t: "TEACHER",
        b: "GROUP",
        //b: "ROOM",
    ),
    Teacher: (
        m: "GROUP",
        t: "SUBJECT",
        b: "TEACHER",
        //b: "ROOM",
    ),
    Room: (
        m: "GROUP",
        t: "SUBJECT",
        b: "TEACHER",
    ),
)

// Document title fallbacks
#let titleFallbacks = (
    Class: "Gesamtstundenplan der Klassen",
    Teacher: "Gesamtstundenplan der Lehrkräfte",
    Room: "Gesamtstundenplan der Räume",
)

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

#let PLAN_AREA_HEIGHT = (PAGE_HEIGHT - PAGE_BORDER.top
    - PAGE_BORDER.bottom - TITLE_HEIGHT)
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

// Type of table ("Class", "Teacher" or "Room")
#let tableType = xdata.TableType
#if tableType == "Class" {
    ROW_HEIGHT = ROW_HEIGHT_CLASS
    V_HEADER_WIDTH = V_HEADER_WIDTH_CLASS
}

// Determine the field placements in the tiles
#let fieldPlacements = typstMap.at("FieldPlacements", default: (:))
#if fieldPlacements.len() == 0 {
    // fallback
    fieldPlacements = boxText.at(tableType, default: (:))
}

// +++ Set up the table
#let nhours = HOURS.len()
#let pcols = DAYS.len() * nhours
#let colwidth = (PLAN_AREA_WIDTH - V_HEADER_WIDTH) / pcols

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
//TODO: Maybe the vertical headers should be boxed, to have auto-adjusting size?

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

#let cell_inset = CELL_BORDER
#let cell_width = colwidth - cell_inset * 2
#let cell_height = ROW_HEIGHT - cell_inset * 2

// Used by fgWCAG2
#let rgblumin(c) = {
    c = c / 100%
    if c <= 0.04045 {
        c/12.92
    } else {
        calc.pow((c+0.055)/1.055, 2.4)
    }
}

// Decide on black or white for text, based on background colour (WCAG2).
#let fgWCAG2(colour) = {
    let (r,g,b,a) = rgb(colour).components()
    let l = 0.2126 * rgblumin(r) + 0.7152 * rgblumin(g) + 0.0722 * rgblumin(b)
    if l > 0.179 { black } else { white }
}

// Used by fgWCAG2
#let rgblumin(c) = {
    c = c / 100%
    if c <= 0.04045 {
        c/12.92
    } else {
        calc.pow((c+0.055)/1.055, 2.4)
    }
}

// Decide on black or white for text, based on background colour (WCAG2).
#let fgWCAG2(colour) = {
    let (r,g,b,a) = rgb(colour).components()
    let l = 0.2126 * rgblumin(r) + 0.7152 * rgblumin(g) + 0.0722 * rgblumin(b)
    if l > 0.179 { black } else { white }
}

// Prepare horizontal header, also column sizes and boundaries
#let dheader = ([],)
#let pheader = ([],)
#let vlines = (V_HEADER_WIDTH,)
#let x = V_HEADER_WIDTH
#for d in DAYS {
    dheader.push(table.cell(colspan: nhours, d))
    for p in HOURS {
        pheader.push(p)
        x += colwidth
        vlines.push(x)
    }
}
#let tcolumns = (V_HEADER_WIDTH,) + (colwidth,)*pcols
#let hcellfill = ([],) * pcols

// Prepare vertical header and row sizes and boundaries
#let nprows = int(TABLE_BODY_HEIGHT / ROW_HEIGHT)
#let trows = ((H_HEADER_HEIGHT1, H_HEADER_HEIGHT2) + (ROW_HEIGHT,)*nprows)
#let hlines = (H_HEADER_HEIGHT,)
#let y = H_HEADER_HEIGHT
#let i = 0
#while i < nprows {
    i += 1
    y += ROW_HEIGHT
    hlines.push(y)
}

#let ttvcell(
    row,
    day: 0,
    hour: 0,
    duration: 1,
    offset: 0,
    fraction: 1,
    total: 1,
    subject: "",
    groups: (),
    teachers: (),
    rooms: (),
    background: "",
) = {
    // Determine grid lines
    let ix = day * nhours + hour
    let x0 = vlines.at(ix)
    let y0 = hlines.at(row)
    // Prepare texts
    let texts = (
        SUBJECT: subject,
        GROUP: groups.join(JOINSTR),
        TEACHER: teachers.join(JOINSTR),
        ROOM: rooms.join(JOINSTR),
    )
    let ctext = texts.at(fieldPlacements.at("m", default: ""), default: "")
    let ttext = texts.at(fieldPlacements.at("t", default: ""), default: "")
    let btext = texts.at(fieldPlacements.at("b", default: ""), default: "")
    let cellBorderColour = background
    if background == "" {
        background = "#FFFFFF"
        cellBorderColour="#000000"
    }
    let bg = rgb(background)
    // Get text colour
    //TODO: choose algorithm for text colour.
    // 1) This converts the background to grey-scale and uses a threshold:
    let bw = luma(bg)
    let textcolour = if bw.components().at(0) < 55% { white } else { black }
    // 2) This uses the WCAG2 guidelines:
    //let textcolour = fgWCAG2(bg)
    set text(textcolour)

    // Determine size and offset of tile
    let w = colwidth * duration - cell_inset * 2
    let hfrac = cell_height * fraction / total
    let yshift = cell_height * offset / total
    // Shrink excessively large components.
    let b = box(
        fill: bg,
        stroke: (paint: rgb(cellBorderColour),thickness:CELL_BORDER),
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

#show heading: it => text(weight: "bold", size: BIG_SIZE,
    bottom-edge: "descender",
    pad(left: 0mm, it))

// Determine the document title
#let doctitle = typstMap.at("Title", default: "")
#if doctitle == "" {
    // fallback
    doctitle = titleFallbacks.at(tableType, default: "")
}
#let subtitle = typstMap.at("Subtitle", default: "")
#let lastChange = typstMap.at("LastChange", default: "")
#if subtitle == "" {
    subtitle = lastChange
} else if lastChange != "" {
    subtitle += " | " + lastChange
}

#set page(height: PAGE_HEIGHT, width: PAGE_WIDTH,
    //numbering: "1/1",
    margin: PAGE_BORDER,
)

// +++ Divide the data into pages
#let irow = 0
#let rows = xdata.Pages
#let nrows = rows.len()
#let xrows = ()
#let iy = 0
#let aix = 0
// Count page numbers
#let pageno = 1
#let pagetotal = int((rows.len() + nprows - 1) / nprows)
#while irow < nrows {
    let row = rows.at(irow)
    irow += 1

    //TODO: Which page field (Name or Short)?

    let rh = row.Name
    if rh == "" {
        rh = row.Short
    }
    xrows += (rh,) + hcellfill
    iy += 1
    if iy == nprows or irow == nrows {
        // Page done

        block(height: TITLE_HEIGHT, above: 0mm, below: 0mm, inset: 2mm)[
            #place(top)[= #doctitle]
            #place(top)[#h(1fr) #text(17pt, xdata.Info.Institution)]
            #place(bottom)[
                #subtitle
                #h(1fr)
                #pageno / #pagetotal
            ]
        ]

        box([
            #table(
                columns: tcolumns,
                rows: trows,
                gutter: 0pt,
                stroke: (paint: rgb(FRAME_COLOUR), thickness: 1pt) ,
                inset: 0pt,
                fill: (x, y) =>
                    if y > 1 and x > 0 {
                        rgb(EMPTY_COLOUR)
                    } else {
                        rgb(HEADER_COLOUR)
                    },
                table.header(
                    ..dheader, ..pheader,
                ),
                ..xrows,
            )

            #let rix = 0
            #while aix < irow {
                for a in rows.at(aix).Activities {
                    ttvcell(rix, ..a)
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



