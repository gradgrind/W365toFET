/* This is a script to generate a multiple page document where each class,
 * teacher or room has a page of its own for its weekly timetable.
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
 * Two variants of table are supported:
 *  - plain (default): A simple table is used with each lesson period having
 *                     the same length.
 *  - with breaks:     The height of the lesson periods is proportional to
 *                     their times in minutes. Also the breaks between lessons
 *                     will be shown, proportional to their times. This can
 *                     only work if the lesson periods (hours) are supplied
 *                     with correctly formatted start and end times (hh:mm,
 *                     24-hour clock).
 *                     To select this variant, the option Typst.WithBreaks
 *                     must be true.
 * If lesson period times are supplied, the parameter Typst.WithTimes must be
 * true (default: false) for them to be shown in the period headers.
 */

// To use a different font:
#set text(font: ("Nunito","DejaVu Sans"))
// If the font is not installed on the system, the .ttf or .otf files can be
// placed in "typst_files/_fonts".

#let PAGE_HEIGHT = 210mm
#let PAGE_WIDTH = 297mm
#let PAGE_BORDER = (top:15mm, bottom: 15mm, left: 15mm, right: 15mm)
#let TITLE_HEIGHT = 15mm
#let H_HEADER_HEIGHT = 15mm
#let V_HEADER_WIDTH = 30mm

#set page(height: PAGE_HEIGHT, width: PAGE_WIDTH,
//  numbering: "1",
  margin: PAGE_BORDER,
)
#let CELL_BORDER = 0.5pt
#let BIG_SIZE = 18pt
#let NORMAL_SIZE = 16pt
#let PLAIN_SIZE = 12pt
#let SMALL_SIZE = 12pt

#let FRAME_COLOUR = "#707070"
#let HEADER_COLOUR = "#f0f0f0"
#let EMPTY_COLOUR = "#ffffff"

// Field placement fallbacks
#let boxText = (
    Class: (
        c: "SUBJECT",
        tl: "TEACHER",
        tr: "GROUP",
        //bl: "",
        br: "ROOM",
    ),
    Teacher: (
        c: "GROUP",
        tl: "SUBJECT",
        tr: "TEACHER",
        //bl: "",
        br: "ROOM",
    ),
    Room: (
        c: "GROUP",
        tl: "SUBJECT",
        //tr: "",
        //bl: "",
        br: "TEACHER",
    ),
)

// Page heading fallbacks
#let pageHeadings = (
    Class: "Klasse %S",
    Teacher: "%N (%S)",
    Room: "Raumplan %N (%S)",
)

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

#let PLAN_AREA_HEIGHT = (PAGE_HEIGHT - PAGE_BORDER.top
    - PAGE_BORDER.bottom - TITLE_HEIGHT)
#let PLAN_AREA_WIDTH = (PAGE_WIDTH - PAGE_BORDER.left
    - PAGE_BORDER.right)

#let xdata = json(sys.inputs.ifile)
#let typstMap = xdata.at("Typst", default: (:))

#let DAYS = ()
#for ddata in xdata.Info.Days {
    //TODO: Which field to use
    DAYS.push(ddata.Name)
}

#let HOURS = ()
#let TIMES = ()
#let WITHTIMES = typstMap.at("WithTimes", default: false)
#let WITHBREAKS = typstMap.at("WithBreaks", default: false)

#for hdata in xdata.Info.Hours {
    //TODO: Which field to use
  let hour = hdata.Short
  let time1 = hdata.at("Start")
  let time2 = hdata.at("End")
  if WITHTIMES {
    HOURS.push(hour + "|" + time1 + " – " + time2)
  } else {
    HOURS.push(hour)
  }
  if WITHBREAKS {
    // The period start and end times must be present and correct.
    let (h1, m1) = time1.split(":")
    let (h2, m2) = time2.split(":")
    // Convert the times to minutes only.
    TIMES.push((int(h1) * 60 + int(m1), int(h2) * 60 + int(m2)))
  }
}

// Type of table ("Class", "Teacher" or "Room")
#let tableType = xdata.TableType

// Determine the field placements in the tiles
#let fieldPlacements = typstMap.at("FieldPlacements", default: (:))
#if fieldPlacements.len() == 0 {
    // fallback
    fieldPlacements = boxText.at(tableType, default: (:))
}

#let vfactor = if WITHBREAKS {
  // Here it is a factor with which to multiply the minutes
  let tdelta = TIMES.at(-1).at(1) - TIMES.at(0).at(0)
  (PLAN_AREA_HEIGHT - H_HEADER_HEIGHT) / tdelta
} else {
  // Here it is just the height of a period box
  (PLAN_AREA_HEIGHT - H_HEADER_HEIGHT) / HOURS.len()
}

// Build the row structure
#let table_content = ([],) + DAYS
#let isbreak = (false,)
#let hlines = ()
#let trows = (H_HEADER_HEIGHT,)
#let t = 0mm
#let m0 = -1
#let i = 0
#for h in HOURS {
    if WITHBREAKS {
        let (m1, m2) = TIMES.at(i)
        if m0 < 0 {
            m0 = m1
        }
        let t1 = (m1 - m0) * vfactor
        if t > 0mm {
            trows.push(t1 - t)
            table_content += ("",) + ([],) * DAYS.len()
            isbreak.push(true)
        }
        t = (m2 - m0) * vfactor
        hlines.push((t1 + H_HEADER_HEIGHT, t + H_HEADER_HEIGHT))
        trows.push(t - t1)
    } else {
        let period_height = vfactor
        let t0 = t
        t += period_height
        hlines.push((t0 + H_HEADER_HEIGHT, t + H_HEADER_HEIGHT))
        trows.push(period_height)
    }
    table_content += (h,) + ([],) * DAYS.len()
    isbreak.push(false)
    i += 1
}

// Build the vertical lines

#let vlines = (V_HEADER_WIDTH,)
#let colwidth = (PLAN_AREA_WIDTH - V_HEADER_WIDTH) / DAYS.len()
#let d0 = V_HEADER_WIDTH
//COLWIDTH #colwidth
#for d in DAYS {
    d0 += colwidth
    vlines.push(d0)
}
#let tcolumns = (V_HEADER_WIDTH,) + (colwidth,)*DAYS.len()

#show table.cell: it => {
  if it.y == 0 {
    set text(size: BIG_SIZE, weight: "bold")
    align(center + horizon, it.body.at("text", default: ""))
  } else if it.x == 0 {
    //it.body.fields()
    let txt = it.body.at("text", default: "")
    let t1t2 = txt.split("|")
    let tt = text(size: NORMAL_SIZE, weight: "bold", t1t2.at(0))
    if t1t2.len() > 1 {
        tt += [\ ] + text(size: PLAIN_SIZE, t1t2.at(1))
    }
    align(center + horizon, tt)
  } else {
    it
  }
}

// On text lines (top or bottom of cell) with two text items:
// If one is smaller than 25% of the space, leave this and shrink the
// other to 90% of the reamining space. Otherwise shrink both.
#let scale2inline(vpos, width, text1, text2) = {
    let t1 = text(size: SMALL_SIZE, text1)
    let t2 = text(size: SMALL_SIZE, text2)
    let w4 = width / 4
    context {
        let s1 = measure(t1)
        let s2 = measure(t2)
        if (s1.width + s2.width) > width * 0.9 {
            if s1.width < w4 {
                // shrink only text2
                let w2 = width - s1.width
                let scl = (w2 * 0.9) / s2.width
                place(vpos + left, t1)
                place(vpos + right, text(size: scl * SMALL_SIZE, text2))
            } else if s2.width < w4 {
                // shrink only text1
                let w2 = width - s2.width
                let scl = (w2 * 0.9) / s1.width
                place(vpos + left, text(size: scl * SMALL_SIZE, text1))
                place(vpos + right, t2)
            } else {
                // shrink both
                let scl = (width * 0.9) / (s1.width + s2.width)
                place(vpos + left, text(size: scl * SMALL_SIZE, text1))
                place(vpos + right, text(size: scl * SMALL_SIZE, text2))
            }
        } else {
            place(vpos + left, t1)
            place(vpos + right, t2)
        }
    }
}

#let scaleinline(vpos, width, textc) = {
    let t = text(size: NORMAL_SIZE, weight: "bold", textc)
    context {
        let s = measure(t)
        if s.width > width * 0.9 {
            let scl = (width * 0.9 / s.width)
            let ts = text(size: scl * NORMAL_SIZE, weight: "bold", textc)
            place(vpos + center, ts)
        } else {
            place(vpos + center, t)
        }
    }
}

#let cell_inset = CELL_BORDER
#let cell_width = colwidth - cell_inset * 2

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

#let ttcell(
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
    // Prepare texts
    let texts = (
        SUBJECT: subject,
        GROUP: groups.join(","),
        TEACHER: teachers.join(","),
        ROOM: rooms.join(","),
    )
    let centre = texts.at(fieldPlacements.at("c", default: ""), default: "")
    let tl = texts.at(fieldPlacements.at("tl", default: ""), default: "") 
    let tr = texts.at(fieldPlacements.at("tr", default: ""), default: "") 
    let bl = texts.at(fieldPlacements.at("bl", default: ""), default: "") 
    let br = texts.at(fieldPlacements.at("br", default: ""), default: "") 
	
    let cellBorderColor =background
    if background == "" {
        background = "#FFFFFF"
        cellBorderColor="#000000"
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

    // Determine grid lines
    let (y0, y1) = hlines.at(hour)
    let x0 = vlines.at(day)
    if duration > 1 {
        y1 = hlines.at(hour + duration - 1).at(1)
    }
    let wfrac = cell_width * fraction / total
    let xshift = cell_width * offset / total
    // Shrink excessively large components.
    let b = box(
        fill: rgb(background),
        stroke: (paint: rgb(cellBorderColor), thickness:CELL_BORDER),
        inset: 2pt,
        height: y1 - y0 - CELL_BORDER*2,
        width: wfrac,
    )[
        #scale2inline(top, wfrac, tl, tr)
        #scaleinline(horizon, wfrac, centre)
        #scale2inline(bottom, wfrac, bl, br)
    ]
    place(top + left,
        dx: x0 + CELL_BORDER + xshift,
        dy: y0 + CELL_BORDER,
        b
    )
}

#let tbody = table(
    columns: tcolumns,
    rows: trows,
    gutter: 0pt,
    stroke: rgb(FRAME_COLOUR),
    inset: 0pt,
    fill: (x, y) =>
        if y != 0 {
            if isbreak.at(y) {
                rgb(BREAK_COLOUR)
            } else if x != 0 {
                rgb(EMPTY_COLOUR)
            }
        },
    ..table_content
)

#show heading: it => text(weight: "bold", size: BIG_SIZE,
    bottom-edge: "descender",
    pad(left: 0mm, it))

#let pheadings = typstMap.at("PageHeading", default: (:))
#let phead = pheadings.at(tableType, default: "-")
#if phead == "-" {
    phead = pageHeadings.at(tableType, default: "")
}
#let page = 0
#for p in xdata.Pages {
    if page != 0 {
        pagebreak()
    }
    page += 1

    let title = phead.replace("%N", p.Name).replace("%S", p.Short)
    block(height: TITLE_HEIGHT, above: 0mm, below: 0mm, inset: 2mm)[
        #place(top)[= #title #h(1fr)#text(17pt)[#xdata.Info.Institution]]
       // #place(left + horizon)[]
        #place(bottom)[#typstMap.at("subtitle", default: "")]
    ]

    box([
        #tbody
        #for a in p.Activities {
            ttcell(..a)
        }
    ])
}

