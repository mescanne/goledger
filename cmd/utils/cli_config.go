package utils

import (
	"fmt"
	"net/url"
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
		return fmt.Errorf("invalid config '%s': should be '<type>:arg=value[;arg=value,...]'", config)
	}

	values, err := url.ParseQuery(parts[1])
	if err != nil {
		return fmt.Errorf("invalid config '%s': %v", config, err)
	}

	cfg.Params = make(map[string]string)
	for k, v := range values {
		if len(v) > 1 {
			return fmt.Errorf("invalid config '%s': multiple values for key '%s'", config, k)
		}
		if len(v) == 0 {
			return fmt.Errorf("invalid config '%s': no value for key '%s'", config, k)
		}
		cfg.Params[k] = v[0]
	}

	cfg.ConfigType = parts[0]
	return nil
}

func (e *CLIConfig) String() string {
	if e.ConfigType == "" {
		return ""
	}
	keyv := make([]string, len(e.Params))
	i := 0
	for k, v := range e.Params {
		keyv[i] = url.QueryEscape(k) + "=" + url.QueryEscape(v)
		i++
	}
	return e.ConfigType + ":" + strings.Join(keyv, ";")
}

func (cfg *CLIConfig) Type() string {
	return "<type>:<k=v;...>"
}
