package watcher

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/Scalify/website-content-watcher/pkg/storage"
)

var (
	cleanRegExp *regexp.Regexp
)

func init() {
	var err error
	cleanRegExp, err = regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
}

func (w *Watcher) getValues(jobName string) (map[string]string, error) {
	jobName = w.cleanJobName(jobName)
	valuesStr, err := w.storage.Get(jobName)
	if err != nil && err != storage.ErrNotFound {
		return nil, err
	}

	values := make(map[string]string)
	if valuesStr == "" {
		return values, nil
	}

	if err := json.Unmarshal([]byte(valuesStr), &values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal values: %v", err)
	}

	return values, nil
}

func (w *Watcher) setValues(jobName string, values map[string]string) error {
	jobName = w.cleanJobName(jobName)
	valueBytes, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshall values: %v", err)
	}

	return w.storage.Set(jobName, string(valueBytes))
}

func (w *Watcher) cleanJobName(jobName string) string {
	return cleanRegExp.ReplaceAllString(jobName, "")
}
