Import Formats

The format for the import configuration:

  type:key=value[,key=value,...]

There is one import format type 'csv', but more can be added in the future.

Import type 'csv' Parameters

  payee, date, amount - 0-based column index for the payee, transaction date,
                        and transaction amount
  skip                - number of header lines to skip (default is 0)
  delim               - delimiter for CSV file (default is ,)



