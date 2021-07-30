// Package uitable provides a decorator for formating data as a table
package uitable

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gosuri/uitable/util/strutil"
	"github.com/gosuri/uitable/util/wordwrap"
)

// Separator is the default column seperator
var Separator = "\t"

// Table represents a decorator that renders the data in formatted in a table
type Table struct {
	// Rows is the collection of rows in the table
	Rows []*Row

	// MaxColWidth is the maximum allowed width for cells in the table
	MaxColWidth uint

	// Wrap when set to true wraps the contents of the columns when the length exceeds the MaxColWidth
	Wrap bool

	// Separator is the seperator for columns in the table. Default is "\t"
	Separator string

	mtx        *sync.RWMutex
	rightAlign map[int]bool
}

// New returns a new Table with default values
func New() *Table {
	return &Table{
		Separator:  Separator,
		mtx:        new(sync.RWMutex),
		rightAlign: map[int]bool{},
	}
}

// AddRow adds a new row to the table
func (t *Table) AddRow(data ...interface{}) *Table {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	r := NewRow(data...)
	t.Rows = append(t.Rows, r)
	return t
}

// Bytes returns the []byte value of table
func (t *Table) Bytes() []byte {
	return []byte(t.String())
}

func (t *Table) RightAlign(col int) {
	t.mtx.Lock()
	t.rightAlign[col] = true
	t.mtx.Unlock()
}

// String returns the string value of table
func (t *Table) String() string {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	if len(t.Rows) == 0 {
		return ""
	}

	// determine the width for each column (cell in a row)
	var colwidths []uint
	for _, row := range t.Rows {
		for i, cell := range row.Cells {
			// resize colwidth array
			if i+1 > len(colwidths) {
				colwidths = append(colwidths, 0)
			}
			cellwidth := cell.LineWidth()
			if t.MaxColWidth != 0 && cellwidth > t.MaxColWidth {
				cellwidth = t.MaxColWidth
			}

			if cellwidth > colwidths[i] {
				colwidths[i] = cellwidth
			}
		}
	}

	var lines []string
	for _, row := range t.Rows {
		row.Separator = t.Separator
		for i, cell := range row.Cells {
			cell.Width = colwidths[i]
			cell.Wrap = t.Wrap
			cell.RightAlign = t.rightAlign[i]
		}
		lines = append(lines, row.String())
	}
	return strutil.Join(lines, "\n")
}

// Row represents a row in a table
type Row struct {
	// Cells is the group of cell for the row
	Cells []*Cell

	// Separator for tabular columns
	Separator string
}

// NewRow returns a new Row and adds the data to the row
func NewRow(data ...interface{}) *Row {
	r := &Row{Cells: make([]*Cell, len(data))}
	for i, d := range data {
		r.Cells[i] = &Cell{Data: d}
	}
	return r
}

// String returns the string representation of the row
func (r *Row) String() string {
	// get the max number of lines for each cell
	var lc int // line count
	for _, cell := range r.Cells {
		if clc := len(strings.Split(cell.String(), "\n")); clc > lc {
			lc = clc
		}
	}

	// allocate a two-dimentional array of cells for each line and add size them
	cells := make([][]*Cell, lc)
	for x := 0; x < lc; x++ {
		cells[x] = make([]*Cell, len(r.Cells))
		for y := 0; y < len(r.Cells); y++ {
			cells[x][y] = &Cell{Width: r.Cells[y].Width}
		}
	}

	// insert each line in a cell as new cell in the cells array
	for y, cell := range r.Cells {
		lines := strings.Split(cell.String(), "\n")
		for x, line := range lines {
			cells[x][y].Data = line
		}
	}

	// format each line
	lines := make([]string, lc)
	for x := range lines {
		line := make([]string, len(cells[x]))
		for y := range cells[x] {
			line[y] = cells[x][y].String()
		}
		lines[x] = strutil.Join(line, r.Separator)
	}
	return strings.Join(lines, "\n")
}

// Cell represents a column in a row
type Cell struct {
	// Width is the width of the cell
	Width uint

	// Wrap when true wraps the contents of the cell when the lenght exceeds the width
	Wrap bool

	// RightAlign when true aligns contents to the right
	RightAlign bool

	// Data is the cell data
	Data interface{}
}

// LineWidth returns the max width of all the lines in a cell
func (c *Cell) LineWidth() uint {
	width := 0
	for _, s := range strings.Split(c.String(), "\n") {
		w := strutil.StringWidth(s)
		if w > width {
			width = w
		}
	}
	return uint(width)
}

// String returns the string formated representation of the cell
func (c *Cell) String() string {
	if c.Data == nil {
		return strutil.PadLeft(" ", int(c.Width), ' ')
	}
	col := color.New(color.FgBlack)
	col.DisableColor()
	s := fmt.Sprintf("%v", col.Sprint(c.Data))
	if c.Width > 0 {
		if c.Wrap && uint(len(s)) > c.Width {
			return wordwrap.WrapString(s, c.Width)
		} else {
			return strutil.Resize(s, c.Width, c.RightAlign)
		}
	}
	return s
}
