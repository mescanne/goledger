package script

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go.starlark.net/starlark"
	"io"
	"math/big"
)

type StarlarkReader func(r io.Reader) (starlark.Value, error)

func jsonToStarlark(data interface{}) (starlark.Value, error) {
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
			newarr[i], err = jsonToStarlark(nv)
			if err != nil {
				return nil, err
			}
		}
		return starlark.NewList(newarr), nil
	case map[string]interface{}:
		newdict := starlark.NewDict(len(v))
		for k, nv := range v {
			sv, err := jsonToStarlark(nv)
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
}

// This is a Starlark Reader
func ReadJSON(r io.Reader) (starlark.Value, error) {
	var p interface{}
	err := json.NewDecoder(r).Decode(&p)
	if err != nil {
		return nil, fmt.Errorf("json parsing: %w", err)
	}

	s, err := jsonToStarlark(p)
	if err != nil {
		return nil, fmt.Errorf("json conversion: %w", err)
	}

	return s, nil
}

func GetReadCSV(delim string, header bool) StarlarkReader {

	return func(r io.Reader) (starlark.Value, error) {
		var err error

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

		// Read all
		recs, err := csvr.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("csv body: %w", err)
		}

		// Convert all
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
	}
}

func GetBigRat(amount starlark.Value) (*big.Rat, error) {
	if amount == starlark.None {
		return nil, nil
	}

	var amt *big.Rat = &big.Rat{}
	switch v := amount.(type) {
	case starlark.Float:
		amt.SetFloat64(float64(v))
	case starlark.Int:
		amt.SetInt(v.BigInt())
	case starlark.String:
		_, ok := amt.SetString(string(v))
		if !ok {
			return nil, fmt.Errorf("invalid amount '%s'", string(v))
		}
	default:
		return nil, fmt.Errorf("not a valid amount type %T (string, int, float)", amount)
	}

	return amt, nil
}
