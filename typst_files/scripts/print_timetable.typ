/* The basic idea is to use a table for structuring each timetable. This
 * manages the headers, lines and background colouring.
 * The tiles of the actual timetable are overlaid on this using the line
 * coordinates for orientation.
 * A space (block) is left at the top of each page for a page title. This
 * is in the space left free by the TITLE_HEIGHT value.
 * The rest of the page will be used for the table, adjusting the cell size
 * to fit.
 * Two variants of table are supported:
 *  - plain (default): A simple table is used with each lesson period having
 *                     the same length.
 *  - with breaks:     The height of the lesson periods is proportional to
 *                     their times in minutes. Also the breaks between lessons
 *                     will be shown, proportional to their times. This can
 *                     only work if the lesson periods (hours) are supplied
 *                     with correctly formatted start and end times (hh:mm,
 *                     24-hour clock).
 *                     To select this variant, the parameter Info.WithBreaks
 *                     must be true.
 * If lesson period times are supplied, the parameter Info.WithTimes must be
 * true (default: false) for them to be shown in the period headers.
 */

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
#let CELL_BORDER = 1pt
#let BIG_SIZE = 16pt
#let NORMAL_SIZE = 14pt
#let PLAIN_SIZE = 12pt
#let SMALL_SIZE = 10pt

#let FRAME_COLOUR = "#707070"
#let BREAK_COLOUR = "#e0e0e0"
#let EMPTY_COLOUR = "#f0f0f0"

#let PLAN_AREA_HEIGHT = (PAGE_HEIGHT - PAGE_BORDER.top
    - PAGE_BORDER.bottom - TITLE_HEIGHT)
#let PLAN_AREA_WIDTH = (PAGE_WIDTH - PAGE_BORDER.left
    - PAGE_BORDER.right)

//#PLAN_AREA_WIDTH x #PLAN_AREA_HEIGHT
#let xdata = json(sys.inputs.ifile)

#let DAYS = xdata.Info.Days
//#let DAYS = ("Mo", "Di", "Mi", "Do", "Fr")
#let HOURS = ()
#let TIMES = ()

#let WITHTIMES = xdata.Info.at("WithTimes", default: false)
#let WITHBREAKS = xdata.Info.at("WithBreaks", default: false)
//#let WITHTIMES = true
//#let WITHBREAKS = true

#for hdata in xdata.Info.Hours {
  let hour = hdata.at("Hour")
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

#let vfactor = if WITHBREAKS {
  // Here it is a factor with which to multiply the minutes
  let tdelta = TIMES.at(-1).at(1) - TIMES.at(0).at(0)
  (PLAN_AREA_HEIGHT - H_HEADER_HEIGHT) / tdelta
} else {
  // Here it is just the height of a period box
  (PLAN_AREA_HEIGHT - H_HEADER_HEIGHT) / HOURS.len()
}
//#vfactor

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

//#trows
//#hlines

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
//#tcolumns
//#vlines

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
#let fit2inspace(width, text1, text2) = {
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
                box(width: width, inset: 2pt,
                    t1
                    + h(1fr)
                    + text(size: scl * SMALL_SIZE, text2)
                )
            } else if s2.width < w4 {
                // shrink only text1
                let w2 = width - s2.width
                let scl = (w2 * 0.9) / s1.width
                box(width: width, inset: 2pt,
                    text(size: scl * SMALL_SIZE, text1)
                    + h(1fr)
                    + t2
                )
            } else {
                // shrink both
                let scl = (width * 0.9) / (s1.width + s2.width)
                box(width: width, inset: 2pt,
                    text(size: scl * SMALL_SIZE, text1)
                    + h(1fr)
                    + text(size: scl * SMALL_SIZE, text2)
                )
            }
        } else {
            box(width: width, inset: 2pt, t1 + h(1fr) + t2)
        }
    }
}

#let fitinspace(width, textc) = {
    let t = text(size: NORMAL_SIZE, weight: "bold", textc)
    context {
        let s = measure(t)
        if s.width > width * 0.9 {
            let scl = (width * 0.9 / s.width)
            let ts = text(size: scl * NORMAL_SIZE, weight: "bold", textc)
            box(width: width, h(1fr) + ts + h(1fr))
        } else {
            box(width: width, h(1fr) + t + h(1fr))
        }
    }
}

#let cell_inset = CELL_BORDER
#let cell_width = colwidth - cell_inset * 2

#let ttcell(
    day: 0,
    hour: 0,
    duration: 1,
    offset: 0,
    fraction: 1,
    total: 1,
    centre: "",
    tl: "",
    tr: "",
    bl: "",
    br: "",
) = {
    let (y0, y1) = hlines.at(hour)
    let x0 = vlines.at(day)
    if duration > 1 {
        y1 = hlines.at(hour + duration - 1).at(1)
    }
    let wfrac = cell_width * fraction / total
    let xshift = cell_width * offset / total
    // Shrink excessively large components.
    let b = box(
        fill: luma(100%),
        stroke: CELL_BORDER,
        inset: 0pt,
        height: y1 - y0 - CELL_BORDER*2,
        width: wfrac,
    )[
        #fit2inspace(wfrac, tl, tr)
        #v(1fr)
        #fitinspace(wfrac, centre)
        #v(1fr)
        #fit2inspace(wfrac, bl, br)
    ]
    place(top + left,
        dx: x0 + CELL_BORDER + xshift,
        dy: y0 + CELL_BORDER,
        b
    )
}

//#context here().position()
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
//  align: center + horizon,
    ..table_content
)

#show heading: it => text(weight: "bold", size: BIG_SIZE,
    bottom-edge: "descender",
    pad(left: 5mm, it))

#let page = 0
#for (k, kdata) in xdata.Pages [
    #{
        if page != 0 {
            pagebreak()
        }
        page += 1
    }

    #block(height: TITLE_HEIGHT, above: 0mm, below: 0mm)[
        #v(5mm)
        = #k
    ]

    #box([
        #tbody
        #for kd in kdata {
            ttcell(..kd)
        }
    ])
]
