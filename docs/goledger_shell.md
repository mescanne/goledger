## goledger shell

Shell integration

### Synopsis

Shell integration

This integrates into the shell. For bash for example:

  eval "$(goledger shell --type=bash)"

This is designed to make it much easier to use goledger
from the command line.


```
goledger shell [flags]
```

### Options

```
  -h, --help         help for shell
      --type shell   Shell for integration (values bash, zsh, powershell)
```

### Options inherited from parent commands

```
      --ccy string       base currency
      --colour           colour (ansi) for reports (default true)
      --divider string   divider for account components for reports (default ":")
      --lang string      language (default "en_GB.UTF-8")
  -l, --ledger string    ledger to read (default "main.ledger")
      --verbose          verbose
```

### SEE ALSO

* [goledger](goledger.md)	 - goledger text-based account application

