package slack

import (
	"encoding/json"
	"gopkg.in/djherbis/times.v1"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const cacheDir = "./.terraform/plugins/.cache/terraform-provider-slack"

func saveCacheAsJson(name string, v interface{}) {
	_ = os.MkdirAll(cacheDir, 0755)
	cacheFile := strings.Join([]string{cacheDir, name}, string(os.PathSeparator))

	if cache, err := json.Marshal(v); err == nil {
		_ = ioutil.WriteFile(cacheFile, cache, 0644)
	} // ignore err
}

func restoreJsonCache(name string, v interface{}) bool {
	_ = os.MkdirAll(cacheDir, 0755)
	cacheFile := strings.Join([]string{cacheDir, name}, string(os.PathSeparator))

	// cache active duration is 1 min
	if t, err := times.Stat(cacheFile); err == nil {
		if !time.Now().After(t.ModTime().Add(1 * time.Minute)) {
			if bytes, err := ioutil.ReadFile(cacheFile); err == nil {
				return json.Unmarshal(bytes, v) == nil
			}
		}
	}

	return false
}
