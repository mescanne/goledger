package utils

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ToJson(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", data)
	} else {
		return string(b)
	}
}

func GetStringValue(data interface{}, key string) (string, error) {
	d, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected JSON object, got: %s", ToJson(data))
	}
	id, ok := d[key]
	if !ok {
		return "", fmt.Errorf("expected key %s missing, got: %s", key, ToJson(data))
	}
	i, ok := id.(string)
	if !ok {
		return "", fmt.Errorf("expected string for key %s got %v", key, id)
	}
	return i, nil
}

func LoadFromFile(file string, data interface{}) error {

	// Ensure directory exists
	err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed mkdir %s: %v", filepath.Dir(file), err)
	}

	// Open the file
	var fh io.Reader
	fh, err = os.Open(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed opening %s: %v", file, err)
	}

	// Check if it's .gz
	if filepath.Ext(file) == ".gz" {
		fh, err = gzip.NewReader(fh)
		if err != nil {
			return fmt.Errorf("failed reading gzip %s: %v", file, err)
		}
	}

	// Read the contents
	b, err := ioutil.ReadAll(fh)
	if err != nil {
		return fmt.Errorf("failed reading from %s: %v", file, err)
	}

	// Unmarshal JSON bytes
	err = json.Unmarshal(b, data)
	if err != nil {
		return fmt.Errorf("failed unmarshalling json from %s: %v", file, err)
	}

	return nil
}

func SaveToFile(file string, data interface{}) (oerr error) {

	// Ensure directory exists
	err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed mkdir %s: %v", filepath.Dir(file), err)
	}

	// Open the file
	fh, err := os.OpenFile(file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("failed opening %s: %v", file, err)
	}

	// Defered closing
	defer func() {
		if err := fh.Close(); err != nil {
			if oerr != nil {
				oerr = fmt.Errorf("failed closing %s: %v", file, err)
			}
		}
	}()

	// Check if it's gz
	var w io.Writer = fh
	if filepath.Ext(file) == ".gz" {
		gwriter := gzip.NewWriter(fh)
		defer func() {
			if err := gwriter.Close(); err != nil {
				if oerr != nil {
					oerr = fmt.Errorf("failed closing %s: %v", file, err)
				}
			}
		}()
		w = gwriter
	}

	// Marshal the JSON bytes
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshalling to json %s: %v", file, err)
	}

	// Write the file
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("failed writing to %s: %v", file, err)
	}

	return nil
}
