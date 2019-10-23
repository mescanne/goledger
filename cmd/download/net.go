package download

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var throttler = getThrottler(time.Millisecond * 100)

var forbidden = errors.New("status 403: forbidden")

func getThrottler(min_delay time.Duration) func() {
	lasttime := time.Unix(0, 0)
	return func() {
		ctime := time.Now()
		delay := lasttime.Add(min_delay).Sub(ctime)
		if delay > 0 {
			time.Sleep(delay)
			lasttime = ctime.Add(delay)
		} else {
			lasttime = ctime
		}
	}
}

func fetchFromURL(client *http.Client, url string, data interface{}) error {
	throttler()
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("url %v: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("url %v: %w", url, forbidden)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("url %v: status %v", url, resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading url %v: %w", url, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(bodyBytes))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(data); err != nil {
		return fmt.Errorf("decoding '%v' from url %v: %w", string(bodyBytes), url, err)
	}
	return nil
}

func CaptureCode(port int) (string, string, error) {
	code := ""
	state := ""
	var clientErr error = nil
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		codes, codeOk := vals["code"]
		states, stateOk := vals["state"]
		if !codeOk || len(codes) != 1 || !stateOk || len(states) != 1 {
			clientErr = fmt.Errorf("Query must specify state and code")
		} else {
			code = codes[0]
			state = states[0]
		}
		server.Close()
	})
	fmt.Printf("Serving is waiting for response...")
	err := server.ListenAndServe()
	if err != http.ErrServerClosed {
		fmt.Printf("Returning error %v\n", err)
		return "", "", err
	}
	if clientErr != nil {
		fmt.Printf("Returning client error %v\n", clientErr)
		return "", "", clientErr
	}
	fmt.Printf("Returning code %v, state %v, no error\n", code, state)
	return code, state, nil
}
