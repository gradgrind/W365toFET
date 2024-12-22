/* Print the complete timetable for classes, teachers or rooms in a fairly
 * compact form.
*/

// To use a different font:
//#set text(font: "B612")
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
#let ROW_HEIGHT = 12mm
#let ROW_HEIGHT_CLASS = 30mm // larger because of divisions

#let CELL_BORDER = 1pt
#let BIG_SIZE = 14pt
#let NORMAL_SIZE = 12pt
#let CELL_TEXT_SIZE = 10pt
#let DAY_SIZE = 13pt
#let HOUR_SIZE = 9pt

#let FRAME_COLOUR = "#707070"
#let HEADER_COLOUR = "#e0e0e0"
#let EMPTY_COLOUR = "#f0f0f0"

// Field placement fallbacks
#let boxText = (
    Class: (
        c: "SUBJECT",
        t: "TEACHER",
        b: "GROUP",
        //b: "ROOM",
    ),
    Teacher: (
        c: "GROUP",
        t: "SUBJECT",
        b: "TEACHER",
        //b: "ROOM",
    ),
    Room: (
        c: "GROUP",
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
//#PLAN_AREA_WIDTH x #PLAN_AREA_HEIGHT

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
        let t = text(size: tsize, weight: wt, textc)
        let s = measure(t)
        if s.width > width * 0.9 {
            let scl = (width * 0.9 / s.width)
            t = text(size: scl * tsize, weight: wt, textc)
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

//#tcolumns
//#vlines

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
//#trows
//#hlines

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
        GROUP: groups.join(","),
        TEACHER: teachers.join(","),
        ROOM: rooms.join(","),
    )
    let ctext = texts.at(fieldPlacements.at("c", default: ""), default: "")
    let ttext = texts.at(fieldPlacements.at("t", default: ""), default: "") 
    let btext = texts.at(fieldPlacements.at("b", default: ""), default: "") 

    if background == "" {
        background = "#FFFFFF"
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
        fill: rgb(background),
        stroke: CELL_BORDER,
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
    pad(left: 5mm, it))

// Determine the document title
#let titles = typstMap.at("Titles", default: (:))
#if titles.len() == 0 {
    // fallback
    titles = titleFallbacks
}
#let title = titles.at(tableType, default: "")
#let subtitle = typstMap.at("Subtitle", default: "")
#let foot1 = [*#title*]
#if subtitle != "" {
    foot1 += [: #subtitle]
}

#set page(height: PAGE_HEIGHT, width: PAGE_WIDTH,
    //numbering: "1/1",
    margin: PAGE_BORDER,
    footer: context [
        #foot1
        #h(1fr)
        #counter(page).display(
            "1/1",
            both: true,
        )
    ]
)

// +++ Divide the data into pages
#let irow = 0
#let rows = xdata.Pages
#let nrows = rows.len()
#let xrows = ()
#let iy = 0
#let aix = 0
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

        //TODO: add page number/of
        
        block(height: TITLE_HEIGHT, above: 0mm, below: 0mm, inset: 2mm)[
            #place(top)[#h(1fr)#xdata.Info.Institution]
            #place(left + horizon)[= #title]
            #place(bottom)[#h(1fr)#typstMap.at("Subtitle", default: "")]
        ]

        box([
            #table(
                columns: tcolumns,
                rows: trows,
                gutter: 0pt,
                stroke: rgb(FRAME_COLOUR),
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
    }
}
