package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type CLIConfig struct {
	ConfigType string
	Params     map[string]string
}

func (cfg *CLIConfig) GetBoolDefault(key string, dftl bool) (bool, error) {
	s, ok := cfg.Params[key]
	if !ok {
		return dftl, nil
	}
	supper := strings.ToUpper(s)
	if supper == "TRUE" {
		return true, nil
	} else if supper == "FALSE" {
		return false, nil
	}

	return false, fmt.Errorf("invalid bool '%s'", s)
}

func (cfg *CLIConfig) GetIntDefault(key string, dftl int) (int, error) {
	s, ok := cfg.Params[key]
	if !ok {
		return dftl, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer '%s': %v", s, err)
	}
	return i, nil
}

func (cfg *CLIConfig) GetStringDefault(key string, dftl string) string {
	v, ok := cfg.Params[key]
	if !ok {
		return dftl
	}
	return v
}

func (cfg *CLIConfig) Set(config string) error {
	parts := strings.SplitN(config, ":", 2)

	// Type is set, initialise map
	cfg.ConfigType = parts[0]
	cfg.Params = make(map[string]string)

	if len(parts) != 2 {
		return nil
	}

	// Extract the parameters
	for _, arg := range strings.Split(parts[1], ",") {

		argpair := strings.SplitN(arg, "=", 2)
		if len(argpair) != 2 {
			cfg.Params[argpair[0]] = "true"
			continue
		}

		_, ok := cfg.Params[argpair[0]]
		if ok {
			return fmt.Errorf("invalid config '%s': arg '%s' set more than once", config, argpair[0])
		}

		cfg.Params[argpair[0]] = argpair[1]
	}

	return nil
}

func (e *CLIConfig) String() string {
	if e.ConfigType == "" {
		return ""
	}

	if len(e.Params) == 0 {
		return e.ConfigType
	}

	keyv := make([]string, len(e.Params))
	i := 0
	for k, v := range e.Params {
		keyv[i] = k + "=" + v
		i++
	}

	return e.ConfigType + ":" + strings.Join(keyv, ",")
}

func (cfg *CLIConfig) Type() string {
	return "<type>[:<k=v,...>]"
}
