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

func (cfg *CLIConfig) GetInt(key string) (int, error) {
	s, err := cfg.GetString(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer '%s': %v", s, err)
	}
	return i, nil
}

func (cfg *CLIConfig) GetIntDefault(key string, dftl int) int {
	i, err := cfg.GetInt(key)
	if err != nil {
		return dftl
	}
	return i
}

func (cfg *CLIConfig) GetString(key string) (string, error) {
	v, ok := cfg.Params[key]
	if !ok {
		return "", fmt.Errorf("missing config for '%s'", key)
	}
	return v, nil
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
	if len(parts) != 2 {
		return fmt.Errorf("invalid config '%s': should be '<type>:arg=value[,arg=value,...]'", config)
	}

	// Type is set
	cfg.ConfigType = parts[0]

	// Extract the parameters
	cfg.Params = make(map[string]string)
	for _, arg := range strings.Split(parts[1], ",") {

		argpair := strings.SplitN(arg, "=", 2)
		if len(argpair) != 2 {
			return fmt.Errorf("invalid config '%s': invalid argument '%s', should be arg=value", config, arg)
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
	keyv := make([]string, len(e.Params))
	i := 0
	for k, v := range e.Params {
		keyv[i] = k + "=" + v
		i++
	}
	return e.ConfigType + ":" + strings.Join(keyv, ",")
}

func (cfg *CLIConfig) Type() string {
	return "<type>:<k=v,...>"
}
