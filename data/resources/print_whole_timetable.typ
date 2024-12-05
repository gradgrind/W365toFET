#let PAGE_HEIGHT = 297mm
#let PAGE_WIDTH = 420mm
#let PAGE_BORDER = (top:15mm, bottom: 15mm, left: 15mm, right: 15mm)
#let NORMAL_SIZE = 12pt
#let TITLE_SIZE = 14pt
#let CELL_BORDER = 1pt
#let CELL_TEXT_SIZE = 10pt
#let DAY_SIZE = 13pt
#let HOUR_SIZE = 9pt

#let FRAME_COLOUR = "#707070"
#let HEADER_COLOUR = "#e0e0e0"
#let EMPTY_COLOUR = "#f0f0f0"

#let TITLE_HEIGHT = 20mm
#let PLAN_AREA_HEIGHT = (PAGE_HEIGHT - PAGE_BORDER.top
    - PAGE_BORDER.bottom - TITLE_HEIGHT)
#let PLAN_AREA_WIDTH = (PAGE_WIDTH - PAGE_BORDER.left
    - PAGE_BORDER.right)

//#PLAN_AREA_WIDTH x #PLAN_AREA_HEIGHT

#let DAYS = ("Mo", "Di", "Mi", "Do", "Fr")
#let HOURS = ("HU 1", "HU 2",
    "FS 1", "FS 2", "FS 3", "FS 4", "FS 5", "FS 6", "FS 7")

#let H_HEADER_HEIGHT1 = 10mm
#let H_HEADER_HEIGHT2 = 10mm
#let H_HEADER_HEIGHT = H_HEADER_HEIGHT1 + H_HEADER_HEIGHT2
#let V_HEADER_WIDTH = 30mm
#let ROW_HEIGHT = 10mm

// Build the vertical lines
#let vlines = (V_HEADER_WIDTH,)
#let pcols = DAYS.len()*HOURS.len()
#let colwidth = (PLAN_AREA_WIDTH - V_HEADER_WIDTH) / pcols
#let tcolumns = (V_HEADER_WIDTH,) + (colwidth,)*pcols

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
    tsize: CELL_TEXT_SIZE, 
    bold: false, 
    align: -1,
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
        if align < 0 {
            box(width: width, t + h(1fr))
        } else if align == 0 {
            box(width: width, h(1fr) + t + h(1fr))
        } else {
            box(width: width, h(1fr) + t)
        }
    }
}

#let cell_inset = CELL_BORDER
#let cell_width = colwidth - cell_inset * 2

// This version only caters for full cells (no subdivision) and fixes the
// structure within the cell.
#let ttvcell(
    duration: 1,
    top: "",
    middle: "",
    bottom: "",
) = {
    let w = colwidth * duration - cell_inset * 2
    let b = box(
        fill: luma(100%),
        height: ROW_HEIGHT - CELL_BORDER*2,
        width: w,
    )[
        #shrinkwrap(w, top, align: 1)
        #v(1fr)
        #shrinkwrap(w, middle, bold: true)
        #v(1fr)
        #shrinkwrap(w, bottom, align: 1)
    ]
    table.cell(colspan: duration, b)
}

#let dheader = ([],)
#let pheader = ([],)
#for d in DAYS {
    dheader.push(table.cell(colspan: HOURS.len(), d))
    for p in HOURS {
        pheader.push(p)
    }
}

#show heading: it => text(weight: "bold", size: TITLE_SIZE,
    bottom-edge: "descender",
    pad(left: 5mm, it))

// Test data:
#let xdata = (
    "Title": "Räume – Gesamtansicht",
    "Rows": (
        ("Header": "First Room", "Items": ()),
        ("Header": "Another Room", "Items": (
            (   "day": 1,
                "hour": 2,
                "data": (
                    "duration": 1, 
                    "top":  "Fr", 
                    "middle": "10.A +", 
                    "bottom": "ABC +",
                ),
            ),
        )),
        ("Header": "A Very, Very Long Room", "Items": (
            (   "day": 2,
                "hour": 4,
                "data": (
                    "duration": 2, 
                    "top": "Ma", 
                    "middle": "10.R", 
                    "bottom": "MN",
                ),
            ),
        )),
        ("Header": "Last Room", "Items": ()),
    )
)

#let ijson = sys.inputs.at("ifile", default: "")
#if  ijson != "" {
    xdata = json(ijson)
}
//#let xdata = json(sys.inputs.ifile)

//TODO: Use data to perform some setting up actions (e.g. days and periods)?

#set page(height: PAGE_HEIGHT, width: PAGE_WIDTH,
  margin: PAGE_BORDER,
  footer: context [
    *#xdata.Title*
    #h(1fr)
    #counter(page).display(
      "1/1",
      both: true,
    )
  ]
)

= #xdata.Title

#let xrows = ()
#for row in xdata.Rows {
    let newrow = ([],)*pcols
    let excess = ()
    for item in row.Items {
        let i = item.day * HOURS.len() + item.hour
        let n = item.data.duration
        while n > 1 {
            n -= 1
            excess.push(i + n)
        }
        newrow.at(i) = ttvcell(..item.data)
    }
    if excess.len() != 0 {
        let xs = excess.sorted()
        while xs.len() != 0 {
            newrow.remove(xs.pop())
        }
    }
    xrows += (row.Header,) + newrow
}

#let trows = (
    (H_HEADER_HEIGHT1, H_HEADER_HEIGHT2)
    + (ROW_HEIGHT,)*xdata.Rows.len()
)

#table(
    columns: tcolumns,
    rows: trows,
    gutter: 0pt,
    stroke: rgb(FRAME_COLOUR),
    inset: 1pt,
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
