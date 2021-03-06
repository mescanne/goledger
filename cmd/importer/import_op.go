package importer

import (
	"fmt"
	"github.com/mescanne/goledger/cmd/utils"
	"github.com/mescanne/goledger/script"
)

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

    add(date, desc,
        amt, ccy=ccy, denom=1, account=account,
	amt2=None, ccy2="", denom2=1, account2=None,
	caccount=caccount, note="", lnote="")

Parameters:

    * date. The date as string of the transaction. Expecting YYYY-MM-DD or YYYY/MM/DD.
    * desc. The payee or description of the transaction.
    * amt. The amount of the transaction. This can be a float, integer, or string.
    * ccy. Optional. The currency of the transaction (otherwise use default).
    * denom. Optional. The denominator of the amount (otherwise 1). (Eg for cents it is denom=100).
    * account. Optional. The account for transactions (otherwise use default).
    * amt2, ccy2, denom2, account2. Optional. As primary accounts, but as a secondary posting if needed.
    * caccount. Optional. The counteraccount for all postings (otherwuse use default).
    * note. Optional. The transaction note.
    * lnote. Optional. The posting note.

Example CSV parsing:
    --code "[add(date=r['Date'], desc=r['Description'], amt=r['Amount']) for r in data]"
`

func NewBookImporterByConfig(cfg *utils.CLIConfig) (script.StarlarkReader, error) {
	if cfg.ConfigType == "" {
		return nil, fmt.Errorf("missing import type")
	} else if cfg.ConfigType == "csv" {
		return NewCSVBookImporter(cfg)
	} else if cfg.ConfigType == "json" {
		return script.ReadJSON, nil
	} else {
		return nil, fmt.Errorf("invalid import type: %s", cfg.ConfigType)
	}
}

func NewCSVBookImporter(cfg *utils.CLIConfig) (script.StarlarkReader, error) {
	delim := cfg.GetStringDefault("delim", ",")
	if len([]rune(delim)) != 1 {
		return nil, fmt.Errorf("invalid delimiter '%s': length not one character", delim)
	}

	header, err := cfg.GetBoolDefault("header", false)
	if err != nil {
		return nil, fmt.Errorf("invalid header config: %v", err)
	}

	return script.GetReadCSV(delim, header), nil
}
