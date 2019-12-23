package importer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/mescanne/goledger/cmd/utils"
	"go.starlark.net/starlark"
	"io"
)

type BookImporter func(r io.Reader) (starlark.Value, error)

var ImportUsage = `Import Detailed Help
============================

Format
------

The format for the import configuration:

    type:key=value[,key=value,...]

Configuration file uses "configType" for the type and a
params subsection for key-value.

Import type *csv*

CSV import reads the file as newline-delimited records and
comma-separated (configurable) fields.

Parameters:
  header              - true or false if there is a header (default is false)
                        if there is a header, structure is a dictionary instead of an array
  delim               - delimiter for CSV file (default is ,)

Import type *json*

JSON import reads the entire file as a single JSON structure. No parameters.

Code
----

There needs to be code written in Starlark. This is a subset of Python that is suitable
for embedded script.

Specification can be found [here](https://github.com/google/starlark-go/blob/master/doc/spec.md).

The data structures are loaded into Python-equivalent. JSON is a series of number, string, lists,
and dictionaries. CSV is a list of lists (no header) or list of dictdionaries (with header).

There are four globals in the execution:

    * data. This the parsed data structure.
    * file. The filename that the data came from.
    * error(msg=msg). This function, if called, aborts all operation and reports the err msg.
    * print(...). This function is the same as the normal Python3 print and can be used for debugging.
      It prints out with a semi-colon at the beginning, so is appropriate for adding text to the
      ledger file.
    * add(). See below.

Add is the method used for adding new postings:

    add(date, desc, amt, ccy=ccy, denom=denom, account=account, caccount=caccount, note=note)

Parameters:

    * date. The date as string of the transaction. Expecting YYYY-MM-DD or YYYY/MM/DD.
    * desc. The payee or description of the transaction.
    * amt. The amount of the transaction. This can be a float, integer, or string.
    * ccy. Optional. The currency of the transaction (otherwise use default).
    * denom. Optional. The denominator of the amount (otherwise 1). (Eg for cents it is denom=100).
    * account. Optional. The account for transactions (otherwise use default).
    * caccount. Optional. The counteraccount for transaction (otherwuse use default).
    * note. Optional. The transaction note.
    * lnote. Optional. The posting note.

Example CSV parsing:
    --code "[add(date=r['Date'], desc=r['Description'], amt=r['Amount']) for r in data]"
`

func NewBookImporterByConfig(cfg *utils.CLIConfig) (BookImporter, error) {
	if cfg.ConfigType == "" {
		return nil, fmt.Errorf("missing import type")
	} else if cfg.ConfigType == "csv" {
		return NewCSVBookImporter(cfg)
	} else if cfg.ConfigType == "json" {
		return NewJSONBookImporter(cfg)
	} else {
		return nil, fmt.Errorf("invalid import type: %s", cfg.ConfigType)
	}
}

func NewJSONBookImporter(cfg *utils.CLIConfig) (BookImporter, error) {

	// Starlark value converter for JSON
	var getStarlarkValue func(interface{}) (starlark.Value, error)
	getStarlarkValue = func(data interface{}) (starlark.Value, error) {
		if data == nil {
			return starlark.None, nil
		}

		switch v := data.(type) {
		case bool:
			return starlark.Bool(v), nil
		case float64:
			return starlark.Float(v), nil
		case string:
			return starlark.String(v), nil
		case []interface{}:
			newarr := make([]starlark.Value, len(v))
			for i, nv := range v {
				var err error
				newarr[i], err = getStarlarkValue(nv)
				if err != nil {
					return nil, err
				}
			}
			return starlark.NewList(newarr), nil
		case map[string]interface{}:
			newdict := starlark.NewDict(len(v))
			for k, nv := range v {
				sv, err := getStarlarkValue(nv)
				if err != nil {
					return nil, err
				}
				if sv.Type() == starlark.None.Type() {
					continue
				}
				err = newdict.SetKey(starlark.String(k), sv)
				if err != nil {
					return nil, fmt.Errorf("conversion error - setting key %s to %v", k, nv)
				}
			}
			return newdict, nil
		default:
			return nil, fmt.Errorf("unknown type: %T", v)
		}

		return nil, fmt.Errorf("should never come here")
	}

	// New reader for JSON -> starlark.Value
	return func(r io.Reader) (starlark.Value, error) {
		var p interface{}
		err := json.NewDecoder(r).Decode(&p)
		if err != nil {
			return nil, fmt.Errorf("json parsing: %w", err)
		}

		s, err := getStarlarkValue(p)
		if err != nil {
			return nil, fmt.Errorf("json conversion: %w", err)
		}

		return s, nil
	}, nil
}

func NewCSVBookImporter(cfg *utils.CLIConfig) (BookImporter, error) {

	delim := cfg.GetStringDefault("delim", ",")
	if len([]rune(delim)) != 1 {
		return nil, fmt.Errorf("invalid delimiter '%s': length not one character", delim)
	}

	header, err := cfg.GetBoolDefault("header", false)
	if err != nil {
		return nil, fmt.Errorf("invalid header config: %v", err)
	}

	return func(r io.Reader) (starlark.Value, error) {

		// CSV
		csvr := csv.NewReader(r)
		csvr.Comma = []rune(delim)[0]
		csvr.TrimLeadingSpace = true
		csvr.ReuseRecord = true

		var hdr []string = nil
		if header {
			hdr, err = csvr.Read()
			if err != nil {
				return nil, fmt.Errorf("csv headers: %w", err)
			}
		}

		recs, err := csvr.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("csv body: %w", err)
		}

		nrecs := make([]starlark.Value, len(recs))
		for i, rec := range recs {
			if header {
				if len(rec) > len(hdr) {
					return nil, fmt.Errorf("found more csv columns (%d) than headers (%d) in row", len(rec), len(hdr))
				}
				ndict := starlark.NewDict(len(rec))
				for j, v := range rec {
					ndict.SetKey(starlark.String(hdr[j]), starlark.String(v))
				}
				nrecs[i] = ndict
			} else {
				narr := make([]starlark.Value, len(rec))
				for j, v := range rec {
					narr[j] = starlark.String(v)
				}
				nrecs[i] = starlark.NewList(narr)
			}
		}

		return starlark.NewList(nrecs), nil
	}, nil
}
