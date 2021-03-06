package watcher

import (
	"fmt"
	"strings"

	"github.com/Scalify/website-content-watcher/pkg/api"
	"github.com/Sirupsen/logrus"
	"github.com/robfig/cron"
)

// Watcher watches the content of websites for changes and notifies people.
type Watcher struct {
	logger     *logrus.Entry
	storage    storageClient
	puppet     puppetMasterClient
	notifiers  map[string]notifier
	configFile string
	config     *api.Config
}

// New returns a new watcher instance
func New(logger *logrus.Entry, storage storageClient, puppet puppetMasterClient, configFile string, config *api.Config) *Watcher {
	return &Watcher{
		logger:     logger,
		storage:    storage,
		puppet:     puppet,
		notifiers:  make(map[string]notifier),
		configFile: configFile,
		config:     config,
	}
}

// CheckConfig checks the config for common and/or known mistakes
func (w *Watcher) CheckConfig() error {
	for _, job := range w.config.Jobs {
		for _, not := range job.Notify {
			if _, err := w.getNotifier(not.Type); err != nil {
				return err
			}
		}

		if _, err := cron.Parse(job.Schedule); err != nil {
			return fmt.Errorf("error parsing cron schedule string for job: %v", err)
		}

		if len(strings.TrimSpace(job.Name)) == 0 {
			return fmt.Errorf("empty or invalid job name: %q", job.Name)
		}
	}

	return nil
}

// RunNow executes all jobs instantly and in series
func (w *Watcher) RunNow() error {
	for _, job := range w.config.Jobs {
		if err := w.do(&job); err != nil {
			return err
		}
	}

	return nil
}

// AddNotifier adds a notifier to the list of known notifiers.
func (w *Watcher) AddNotifier(n notifier) error {
	key := n.Key()
	if _, ok := w.notifiers[key]; ok {
		return fmt.Errorf("notifier %q is already registrered", key)
	}

	w.logger.Infof("Added notifier %q to watcher", key)
	w.notifiers[key] = n
	return nil
}

func (w *Watcher) do(job *api.Job) error {
	w.logger.Infof("Running job %s", job.Name)

	pmJob, err := w.executeJob(job)
	if err != nil {
		return fmt.Errorf("failed to execute job %q: %v", job.Name, err)
	}

	if pmJob.Error != "" {
		return fmt.Errorf("job %q of watch job %q execution failed: %v", pmJob.UUID, job.Name, pmJob.Error)
	}

	oldValues, err := w.getValues(job.Name)
	if err != nil {
		return fmt.Errorf("failed to load old values: %v", err)
	}

	newValues := w.transformResults(pmJob.Results)
	diff := diff(newValues, oldValues)
	if err := w.notify(job, diff, newValues); err != nil {
		return err
	}

	w.logger.Infof("Done running job %s", job.Name)

	return w.setValues(job.Name, newValues)
}

func diff(newValues, oldValues map[string]string) []api.Diff {
	diff := make([]api.Diff, 0)

	for key, newVal := range newValues {
		oldVal, ok := oldValues[key]

		if ok && oldVal != "" && oldVal == newVal {
			continue
		}

		diff = append(diff, api.Diff{
			Item:     key,
			OldValue: oldVal,
			NewValue: newVal,
		})
	}

	return diff
}

func (w *Watcher) notify(job *api.Job, diff []api.Diff, newValues map[string]string) error {
	if len(diff) == 0 && job.NotifyOnChangeOnly {
		return nil
	}

	for _, notify := range job.Notify {
		not, err := w.getNotifier(notify.Type)
		if err != nil {
			return err
		}

		if err := not.Notify(job.Name, notify.Value, diff, newValues); err != nil {
			return fmt.Errorf("failed to notify by %q: %v", notify.Type, err)
		}
	}

	return nil
}

func (w *Watcher) getNotifier(name string) (notifier, error) {
	not, ok := w.notifiers[name]
	if !ok {
		return nil, fmt.Errorf("notifier %q not found. It is either not available or not enabled", name)
	}

	return not, nil
}

func (w *Watcher) cronFunc(job *api.Job) func() {
	return func() {
		if err := w.do(job); err != nil {
			w.logger.Error(err)
		}
	}
}

// RegisterCronJobs registers all jobs taken from config at the given cron instance
func (w *Watcher) RegisterCronJobs(cron *cron.Cron) error {
	for i, job := range w.config.Jobs {
		w.logger.Debugf("Adding job %q with pattern %q", job.Name, job.Schedule)
		if err := cron.AddFunc(job.Schedule, w.cronFunc(&(w.config.Jobs[i]))); err != nil {
			return fmt.Errorf("failed to register cron for job %q: %v", job.Name, err)
		}
	}

	return nil
}

func (w *Watcher) transformResults(values map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range values {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}
