package report

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

/*
 * Syntax:
 * columnName: [a-zA-Z]+
 * width: (w[0-9]+)
 * max width: (max[0-9]+)
 * min width: (min[0-9]+)
 * justification: -
 */

/*
 * Main logic:
 * take in a specification, a structure (or slice/array value)
 * Map the specification to the columns.
 * Find the right widths.
 * Do the formatting.
 */

/*
 * First attempt:
 * Get Value:
 * - Struct with a Key with defaults (width, justification, potentially name) are used.
 * - Set validates and turns it into a proper structure.
 *
 * Execute:
 * - Take the above configuration and a slice/array interface.
 * - Iterate and format.
 */

/*
 * Second attempt:
 * Get Value:
 * - Struct with a Key with defaults (width, justification, potentially name) are used.
 * - Set validates and turns it into a proper structure.
 *
 * Extract execute:
 * Take a an array/slice of a structure and turn into a [][]string (array of records)
 *
 * Extract execute:
 * Take a an array/slice of a structure and turn into a [][]string (array of records)
 */

type StringFormat struct {
	Name         string
	Width        int
	JustifyRight bool
}

type Formatter struct {
	Name    string         // Name of the structure
	Formats []StringFormat // Format strings (columns)
}

func parseStringFormat(fmts string) ([]StringFormat, error) {
	o := make([]StringFormat, 0, 10)
	for _, f := range strings.Split(fmts, ",") {
		parts := strings.Split(f, ":")
		width := 0

		if len(parts) > 2 {
			return nil, fmt.Errorf("expected <name>:<width> for column '%s'", f)
		}

		if len(parts) == 2 {
			var err error
			width, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("expected valid integer for <width> '%s'", parts[1])
			}
		}

		o = append(o, StringFormat{
			Name:         strings.TrimSpace(parts[0]),
			Width:        width,
			JustifyRight: false,
		})
	}

	return o, nil
}

//func (fmtter *Formatter) WriteCSV(w io.Writer, v interface{}) {
//	arr, err := appendStruct([]string{}, []string{}, v)
//	// write out arr
//}
//
//func (fmtter *Formatter) WriteText(w io.Writer, v interface{}) {
//
//	// get list of field names.
//	// have header in there if needed.
//	// appendStruct(arr [][]string, fields []string, av interface{}) ([][]string, error) {
//	arr, err := appendStruct([]string{}, []string{}, v)
//
//	// Find the width for each column
//	nwidths := make([]int, len(width), len(width))
//	for j := 0; j < len(width); j++ {
//		colwidth = 0
//		for i := 0; i < len(arr); i++ {
//			ls := Length(arr[i][j], width[j])
//			if ls > colwidth {
//				colwidth = ls
//			}
//			if colwidth == width[j] {
//				break
//			}
//		}
//
//		nwidths[j] = colwidth
//	}
//
//	// Write them out
//	for i := 0; i < len(arr); i++ {
//		for j := 0; j < len(arr[i]); j++ {
//			w.Write(PadString(arr[i][j], nwidths[j], justify_right[j]))
//		}
//		w.Write("\n")
//	}
//}

func formatColumns(arr [][]string, width []int, justify_right []bool) {

}

const ansiStart = '\033'
const ansiEnd = 'm'

func Length(s string, maxcount int) (count int, idx int) {
	count = 0
	idx = 0
	for idx < len(s) && count < maxcount {
		r, width := utf8.DecodeRuneInString(s[idx:])
		idx += width

		// Skip forward if it's ansi
		if r == ansiStart {
			for idx < len(s) && r != ansiEnd {
				r, width = utf8.DecodeRuneInString(s[idx:])
				idx += width
			}

			// Now we have 'm' -- get to the next one
			continue
		}

		// Another character
		count++
	}

	return count, idx
}

func PadString(s string, max int, justify_right bool) string {
	count, idx := Length(s, max)
	if count == max {
		return s[0:idx]
	}

	padding := strings.Repeat(" ", max-count)
	if justify_right {
		return padding + s
	} else {
		return s + padding
	}
}

func appendStruct(header []string, fields []string, av interface{}) ([][]string, error) {

	// Must be array or slice
	v := reflect.ValueOf(av)

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("expected array/slice, got %s", v.Type().String())
	}

	// .. and struct
	if v.Type().Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected array/slice of structure, got %v", v.Type().Elem().String())
	}

	// Iterate formats fields
	sfields := make([]reflect.StructField, len(fields), len(fields))
	for i, f := range fields {
		st, ok := v.Type().Elem().FieldByName(f)
		if !ok {
			return nil, fmt.Errorf("field %s not found in structure %v", f, v.Type().Elem().String())
		}
		sfields[i] = st
	}

	// Initialize capacity on array
	var narr [][]string
	if header != nil {
		narr = make([][]string, 1, v.Len()+1)
		narr[0] = header
	} else {
		narr = make([][]string, 0, v.Len())
	}

	// Need to get a list of string fields.
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		row := make([]string, v.Len(), v.Len())
		for j, st := range sfields {
			row[j] = fmt.Sprintf("%s", e.FieldByIndex(st.Index))
		}
		narr = append(narr, row)
	}

	return narr, nil

}

// Print the array/slice structure
// ... align the width across the array/slice.
func (fmt *Formatter) Print(w io.Writer, i interface{}) {

}

// Print the single structure
func (fmt *Formatter) Printer(w io.Writer, i interface{}) {

}

//
//// Return the current set of fields, min/max width as a nice looking string
//func (fmt *Formatter) String() string {
//	// format is:
//	// %-10.20{columnName}s  or %20.20{columnName}d
//}
//
//func (fmt *Formatter) Positions() []string {
//	// format is:
//	// %-10.20{columnName}s  or %20.20{columnName}d
//}
//
//func (fmt *Formatter) Columns() []string {
//	// Array of columns
//}
//
//// Get the configured string and configure the formatter
//func (fmt *Formatter) Set(v string) error {
//}
//
//// Get the type name
//func (fmt *Formatter) Type() string {
//	return fmt.Name
//}

// *   String() string
// *   Set(string) error
// *   Get() interface{}
// *   Type() string

//func getFormatter(t reflect.Type) (*Formatter, error) {
//	typ := reflect.TypeOf(i)
//	for typ.Kind == reflect.Ptr {
//		typ = typ.Elem()
//	}
//	if typ.Kind == reflect.Struct || reflect.Array {
//		typ = typ.Elem()
//
//	}
//	for typ.Kind == reflect.Ptr {
//		typ = typ.Elem()
//
//	}
//	if typ.Kind != reflect.Struct {
//		return nil, fmt.Errorf("expected structure, or array/slice of structure; got %v", reflect.TypeOf(i).String())
//	}
//
//	// Need to get a list of string fields.
//	fields := make([]string, 0, typ.NumField())
//
//	for i := 0; i < typ.NumField(); i++ {
//		sf := typ.Field(i)
//		cfg, ok := sf.Tag.Lookup("text")
//		if ok {
//			// Add configuration
//			fields = append(fields, sf.Name)
//		}
//	}
//
//	return &Formatter{}, nil
//}

/*
 * Get a structure:
 *  - parse the command line parameter (string)
 *  - generate an fmt.Sprintf() format string
 *  - take a structure and extract an array String
 */

/*
 * Need:
 * - list of column names                   (in struct - or use default)
 * - optional width                         (default in key)
 * - optional justification (left/right)    (default in key)
 */

////
///*
// * Sprintf relationship?
// *
// * We can ->
// * Take a structure and bring out
// *   - parser for formatting string
// *   - generator of Sprintf command line
// *
// * And:
// *   - take a slice of structure
// *   -
// */
//
///*
// * Features:
// * Initiate a "formatter" from a structure.
// * - Return a validation for command line parsing, as well as help text. This is for text or CSV.
// * - Do a CSV output
// * - Do a columnar-text output
// */
//
//type ReportFormatter struct {
//	// Value for output type
//	// Value for column-filtering
//}
//
//type ToString interface {
//	String() string
//}
//
//func getFormatter(t reflect.Type) error {
//	typ := t
//	for typ.Kind == reflect.Ptr {
//		typ = typ.Elem()
//	}
//
//	// Then done
//	if typ.Implements(reflect.typeOf(ToString)) {
//		return func(v reflect.Value) string {
//			v.
//		}
//	}
//
//	switch (typ) {
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64
//		return func(v reflect.Value) string {
//			fmt.Sprintf("%x", v.Int64())
//		}
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
//		return func(v reflect.Value) string {
//			fmt.Sprintf("%x", v.Uint64())
//		}
//	default:
//		return fmt.Errorf("Failed.")
//	//case Bool:
////case reflect.String
////// Not sure
////Float32
////Float64
////Complex64
////Complex128
////Chan
////Func
////Interface
////Map
//////
////Array
////Slice
////// Go deeper
////Struct
////reflect.Uintptr
////UnsafePointer
//	}
//}
//
//func NewReportFormatter(i interface{}) error {
//	typ := reflect.TypeOf(i)
//	if typ.Kind == reflect.Ptr {
//		typ = typ.Elem()
//	}
//	if typ.Kind == reflect.Struct || reflect.Array {
//		typ = typ.Elem()
//
//	}
//	if typ.Kind == reflect.Ptr {
//		typ = typ.Elem()
//
//	}
//	if typ.Kind != reflect.Struct {
//		return fmt.Errorf("expected structure, or array/slice of structure; got %v", reflect.TypeOf(i).String())
//	}
//
//	// Need to get a list of string fields.
//	fields := make([]string, 0, typ.NumField())
//
////	for i := 0; i < typ.NumField(); i++ {
////		sf := typ.Field(i)
////		sf.Name
////		sf.PkgPath
////		sf.Type (type)
////// See through.
////Ptr
////		switch (sf.Type.Kind) {
////			   Bool
////case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64
////case reflect.String
////// Not sure
////Float32
////Float64
////Complex64
////Complex128
////Chan
////Func
////Interface
////Map
//////
////Array
////Slice
////// Go deeper
////Struct
////reflect.Uintptr
////UnsafePointer
////			case
////		}
////		sf.Tag (?)
////		// This is just a STring. But it has:
////		//  Get((key) -> string
////		//  Lookup(key) -> string, ok
////		sf.Index (index for fieldby index?)
////		Anonymous (true/false)
////		// TODO: What about embedded structures? How do we deal with that?
////		// How do we serialize the different types? Just use String()?
//	}
//}
//
///*
// * type Value interface {
// *   String() string
// *   Set(string) error
// *   Get() interface{}
// *   Type() string
// * }
// */
//

//
//// Combine multiple columns into one.
//// Single list length. Right-justified. Space-separated.
//func Combine(strs [][]string, max int) ([]string, error) {
//
//	// Calculate width of each column
//	maxlen := 0
//	for _, str := range strs {
//		if len(str) != len(strs[0]) {
//			return nil, fmt.Errorf("internal error: inconsistent string lengths")
//		}
//		l := ListLength(str, max)
//		if l > maxlen {
//			maxlen = l
//		}
//	}
//
//	// New column
//	ncol := make([]string, len(strs[0]))
//	buf := make([]string, len(strs))
//	for i := range ncol {
//		for j := range strs {
//			buf[j] = PadString(strs[j][i], maxlen, false)
//		}
//		ncol[i] = strings.Join(buf, " ")
//	}
//
//	return ncol, nil
//}
//
///*
// * Sequences using the ESC (escape) character take the form ESC [I...] F, where the ESC character is followed by zero or more intermediate bytes[23] (I) from the range 0x20–0x2F, and one final byte[24] (F) from the range 0x30–0x7E.[25]
// * FRom https://en.wikipedia.org/wiki/ISO/IEC_2022#General_syntax_of_escape_sequences
// */
//const ansiStart = '\033'
//const ansiEnd = 'm'
//
//func Length(s string) int {
//	count := 0
//	idx := 0
//	for idx < len(s) {
//		r, width := utf8.DecodeRuneInString(s[idx:])
//		idx += width
//
//		// Skip forward if it's ansi
//		if r == ansiStart {
//			for idx < len(s) && r != ansiEnd {
//				r, width = utf8.DecodeRuneInString(s[idx:])
//				idx += width
//			}
//
//			// Now we have 'm' -- get to the next one
//			continue
//		}
//
//		// Another character
//		count++
//	}
//
//	return count
//}
//
