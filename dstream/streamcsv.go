package dstream

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

// CSVReader supports reading a Dstream from an io.Reader.
type CSVReader struct {
	rdr    io.Reader
	csvrdr *csv.Reader

	bdata []interface{}

	// Used to hold the first row of data if we needed to read it
	// to get the number of columns.
	stashrec []string

	// If true, skip records with unparseable CSV data, otherwise
	// panic on them.
	skipErrors bool

	comma     rune
	chunkSize int
	nvar      int
	nobs      int
	hasheader bool
	doneinit  bool
	done      bool

	namepos map[string]int
	names   []string

	// If true, all variables are included and converted to float64 type.
	allFloat bool

	// If true, all variables are included and converted to string type.
	allString bool

	// Names of variables to be converted to float64's
	float64Vars []string

	// Names of variables to be stored as strings
	stringVars []string

	// Positions of variables to be converted to floats
	float64VarsPos []int

	// Positions of variables to be stored as strings
	stringVarsPos []int
}

// FromCSV returns a Dstream that reads from a CSV source.  Call at
// least one SetXX method to define variables to be retrieved.  For
// further configuration, chain calls to other SetXXX methods, and
// finally call Done to produce the Dstream.
func FromCSV(r io.Reader) *CSVReader {

	dr := &CSVReader{
		rdr: r,
	}

	return dr
}

// Done is called when all configuration is complete to obtain a Dstream.
func (cs *CSVReader) Done() Dstream {
	cs.init()
	return cs
}

// SkipErrors results in lines with unpareable CSV content being
// skipped (the csv.ParseError is printed to stdio).
func (cs *CSVReader) SkipErrors() *CSVReader {
	cs.skipErrors = true
	return cs
}

// Close does nothing, the caller should explicitly close the
// io.Reader passed to FromCSV if needed.
func (cs *CSVReader) Close() {
}

// HasHeader indicates that the first row of the data file contains
// column names.  The default behavior is that there is no header.
func (cs *CSVReader) HasHeader() *CSVReader {
	if cs.doneinit {
		msg := "FromCSV: can't call HasHeader after beginning data read"
		panic(msg)
	}
	cs.hasheader = true
	return cs
}

// Comma sets the delimiter (comma rune) for the CSVReader.
func (cs *CSVReader) Comma(c rune) *CSVReader {
	cs.comma = c
	return cs
}

// Consistency checks for arguments.
func (cs *CSVReader) checkArgs() {

	if cs.allFloat && cs.allString {
		msg := "Cannot select AllFloat and AllString.\n"
		panic(msg)
	}

	if cs.allFloat && (len(cs.float64Vars) > 0 || len(cs.stringVars) > 0) {
		msg := "Cannot specify AllFloat and FloatVars or StringVars simultaneously"
		panic(msg)
	}

	if cs.allString && (len(cs.float64Vars) > 0 || len(cs.stringVars) > 0) {
		msg := "Cannot specify AllString and FloatVars or StringVars simultaneously"
		panic(msg)
	}
}

func (cs *CSVReader) init() {

	cs.checkArgs()

	if cs.chunkSize == 0 {
		cs.chunkSize = 10000
	}

	cs.csvrdr = csv.NewReader(cs.rdr)
	if cs.comma != 0 {
		cs.csvrdr.Comma = cs.comma
	}

	// Read the first row (may or may not be column header)
	var row1 []string
	var err error
	row1, err = cs.csvrdr.Read()
	if err != nil {
		panic(err)
	}

	// Create a variable name to column index map
	hdrmap := make(map[string]int)
	var vlist []string
	if cs.hasheader {
		for k, v := range row1 {
			hdrmap[v] = k
			vlist = append(vlist, v)
		}
	} else {
		cs.stashrec = row1
		for k := range row1 {
			v := fmt.Sprintf("V%d", k+1)
			hdrmap[v] = k
			vlist = append(vlist, v)
		}
	}

	if cs.allFloat {
		// All variables are selected and have float type
		cs.float64Vars = vlist
	} else if cs.allString {
		// All variables are selected and have float type
		cs.stringVars = vlist
	}

	// Variables to extract must be explicitly selected.
	if len(cs.float64Vars)+len(cs.stringVars) == 0 {
		msg := "No variables specified for reading from CSV file.\n"
		panic(msg)
	}

	// Specify certain variables as having float type
	cs.float64VarsPos = cs.float64VarsPos[0:0]
	for _, v := range cs.float64Vars {
		pos, ok := hdrmap[v]
		if !ok {
			msg := fmt.Sprintf("Variable '%s' not found", v)
			panic(msg)
		}
		cs.float64VarsPos = append(cs.float64VarsPos, pos)
	}

	cs.stringVarsPos = cs.stringVarsPos[0:0]
	for _, v := range cs.stringVars {
		pos, ok := hdrmap[v]
		if !ok {
			msg := fmt.Sprintf("Variable '%s' not found", v)
			panic(msg)
		}
		cs.stringVarsPos = append(cs.stringVarsPos, pos)
	}

	cs.nvar = len(cs.float64Vars) + len(cs.stringVars)
	for _, _ = range cs.float64Vars {
		cs.bdata = append(cs.bdata, make([]float64, 0, 1000))
	}
	for _, _ = range cs.stringVars {
		cs.bdata = append(cs.bdata, make([]string, 0, 1000))
	}

	cs.names = append(cs.float64Vars, cs.stringVars...)

	cs.namepos = make(map[string]int)
	for k, na := range cs.names {
		cs.namepos[na] = k
	}

	cs.doneinit = true
}

// AllFloat results in all variables being selected and converted to
// float64 type.
func (cs *CSVReader) AllFloat64() *CSVReader {
	cs.allFloat = true
	return cs
}

// AllString results in all variables being selected and treated as
// string type.
func (cs *CSVReader) AllString() *CSVReader {
	cs.allString = true
	return cs
}

// SetChunkSize sets the size of chunks for this Dstream, it can only
// be called before reading begins.
func (cs *CSVReader) SetChunkSize(c int) *CSVReader {
	cs.chunkSize = c
	return cs
}

// SetFloatVars sets the names of the variables to be converted to
// float64 type.  Refer to the columns by V1, V2, etc. if there is no
// header row.
func (cs *CSVReader) SetFloat64Vars(x ...string) *CSVReader {
	cs.float64Vars = x
	return cs
}

// SetStringVars sets the names of the variables to be stored as
// string type values.  Refer to the columns by V1, V2, etc. if there
// is no header row.
func (cs *CSVReader) SetStringVars(x ...string) *CSVReader {
	cs.stringVars = x
	return cs
}

// Names returns the names of the variables in the dstream.
func (cs *CSVReader) Names() []string {
	return cs.names
}

// NumVar returns the number of variables in the dstream.
func (cs *CSVReader) NumVar() int {
	return cs.nvar
}

// NumObs returns the number of observations in the dstream.  If the
// dstream has not been fully read, returns -1.
func (cs *CSVReader) NumObs() int {
	if cs.done {
		return cs.nobs
	}
	return -1
}

// GetPos returns a chunk of a data column by column position.
func (cs *CSVReader) GetPos(j int) interface{} {
	return cs.bdata[j]
}

// Get returns a chunk of a data column by name.
func (cs *CSVReader) Get(na string) interface{} {
	pos, ok := cs.namepos[na]
	if !ok {
		msg := fmt.Sprintf("Variable '%s' not found", na)
		panic(msg)
	}
	return cs.bdata[pos]
}

// Reset attempts to reset the Dstream that is reading from an
// io.Reader.  This is only possible if the underlying reader is
// seekable, so reset panics if the seek cannot be performed.
func (cs *CSVReader) Reset() {
	if !cs.doneinit {
		panic("cannot reset, Dstream has not been fully constructed")
	}

	if cs.nobs == 0 {
		return
	}

	r, ok := cs.rdr.(io.ReadSeeker)
	if !ok {
		panic("cannot reset")
	}
	_, err := r.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}
	cs.nobs = 0
	cs.done = false
	cs.rdr = r                   // is this needed?
	cs.csvrdr = csv.NewReader(r) // is this needed?

	// Skip over the header if needed.
	if cs.hasheader {
		_, err := cs.csvrdr.Read()
		if err != nil {
			panic(err)
		}
	}
}

// Next advances to the next chunk.
func (cs *CSVReader) Next() bool {

	if cs.done {
		return false
	}

	truncate(cs.bdata)

	for j := 0; j < cs.chunkSize; j++ {

		var rec []string
		var err error
		if cs.stashrec != nil {
			rec = cs.stashrec
			cs.stashrec = nil
		} else {
			rec, err = cs.csvrdr.Read()
			if err == io.EOF {
				cs.done = true
				return ilen(cs.bdata[0]) > 0
			} else if err != nil {
				if cs.skipErrors {
					os.Stderr.WriteString(fmt.Sprintf("%v\n", err))
					continue
				}
				panic(err)
			}
		}
		cs.nobs++

		for k, pos := range cs.float64VarsPos {
			x, err := strconv.ParseFloat(rec[pos], 64)
			if err != nil {
				x = math.NaN()
			}
			u := cs.bdata[k].([]float64)
			cs.bdata[k] = append(u, x)
		}

		m := len(cs.float64VarsPos)
		for k, pos := range cs.stringVarsPos {
			u := cs.bdata[m+k].([]string)
			cs.bdata[m+k] = append(u, rec[pos])
		}
	}

	return true
}

// csvWriter supports writing a Dstream to an io.Writer in csv format.
type csvWriter struct {

	// The Dstream to be written.
	stream Dstream

	// Format for float type value
	floatFmt string

	// A slice of format types, stored per-variable.
	fmts []string

	wtr io.Writer
}

// ToCSV writes a Dstream in CSV format.  Call SetWriter or Filename
// to configure the underlying writer, then call additional methods
// for customization as desired, and finally call Done to complete the
// writing.
func ToCSV(d Dstream) *csvWriter {
	c := &csvWriter{
		stream: d,
	}
	return c
}

// FloatFmt sets the format string to be used when writing float
// values.  This value is ignored for columns specified in a call to
// the Formats method.
func (dw *csvWriter) FloatFmt(fmt string) *csvWriter {

	dw.floatFmt = fmt
	return dw
}

// Formats sets format strings to be used when writing the Dstream.
// The provided argument is a map from variable names to variable
// formats.
func (dw *csvWriter) Formats(fmts map[string]string) *csvWriter {

	vp := VarPos(dw.stream)

	if dw.fmts == nil {
		nvar := dw.stream.NumVar()
		dw.fmts = make([]string, nvar)
	}
	for v, f := range fmts {
		pos, ok := vp[v]
		if !ok {
			msg := fmt.Sprintf("ToCSV: column %s not found", v)
			panic(msg)
		}
		dw.fmts[pos] = f
	}

	return dw
}

// Filename configures the CSVWriter to write to the given named file.
func (dw *csvWriter) Filename(name string) *csvWriter {

	var err error
	dw.wtr, err = os.Create(name)
	if err != nil {
		panic(err)
	}

	return dw
}

// SetWriter configures the CSVWriter to write to the given io stream.
func (dw *csvWriter) SetWriter(w io.Writer) *csvWriter {

	dw.wtr = w
	return dw
}

// getFmt is a utility for getting the format string for a given
// column.
func (dw *csvWriter) getFmt(t string, col int) string {

	if dw.fmts != nil && dw.fmts[col] != "" {
		return dw.fmts[col]
	}

	switch t {
	case "float":
		if dw.floatFmt == "" {
			return "%.8f"
		} else {
			return dw.floatFmt
		}
	case "int":
		return "%d"
	default:
		panic("unkown type")
	}
}

// Done completes writing a Dstream to a specified io.Writer in csv
// format.
func (dw *csvWriter) Done() error {

	if dw.wtr == nil {
		return errors.New("ToCSV: writer must be set before calling Done")
	}

	csw := csv.NewWriter(dw.wtr)

	err := csw.Write(dw.stream.Names())
	if err != nil {
		return err
	}

	nvar := dw.stream.NumVar()
	rec := make([]string, nvar)
	fmts := make([]string, nvar)

	firstrow := true
	for dw.stream.Next() {
		n := ilen(dw.stream.GetPos(0))

		for i := 0; i < n; i++ {
			for j := 0; j < nvar; j++ {
				// TODO: better support for types
				switch x := dw.stream.GetPos(j).(type) {
				case []float64:
					if firstrow {
						fmts[j] = dw.getFmt("float", j)
					}
					rec[j] = fmt.Sprintf(fmts[j], x[i])
				case []string:
					rec[j] = x[i]
				default:
					rec[j] = ""
				}
			}
			if err := csw.Write(rec); err != nil {
				return err
			}
			firstrow = false
		}
	}

	csw.Flush()

	return nil
}
