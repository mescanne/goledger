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

var ImportFormatUsage = `Import Formats

The format for the import configuration:

  type:key=value[,key=value,...]

There is one import format type 'csv', but more can be added in the future.

Import type 'csv' Parameters

  header              - true or false if there is a header (default is false)
                        if there is a header, structure is a dictionary instead of an array
  delim               - delimiter for CSV file (default is ,)

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
		/* Code for CSV conversion */
		/*
			case []string:
				newarr := make([]starlark.Value, len(v))
				for i, nv := range v {
					newarr[i] = starlark.String(nv)
				}
				return starlark.NewList(newarr), nil
			case [][]string:
				newarr := make([]starlark.Value, len(v))
				for i, nv := range v {
					newarr[i], _ = getStarlarkValue(nv)
				}
				return starlark.NewList(newarr), nil
		*/
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
			return nil, err
		}

		s, err := getStarlarkValue(p)
		if err != nil {
			return nil, err
		}

		return s, err
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
				return nil, err
			}
		}

		recs, err := csvr.ReadAll()
		if err != nil {
			return nil, err
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
				narr := starlark.NewList(make([]starlark.Value, len(recs)))
				for j, v := range rec {
					narr.SetIndex(j, starlark.String(v))
				}
				nrecs[i] = narr
			}
		}

		return starlark.NewList(nrecs), nil
	}, nil
}
