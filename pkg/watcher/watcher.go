package watcher

import (
	"fmt"
	"time"

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
		notifiers:  make(map[string]notifier, 0),
		configFile: configFile,
		config:     config,
	}
}

// AddNotifier adds a notifier to the list of known notifiers.
func (w *Watcher) AddNotifier(n notifier) error {
	key := n.Key()
	if _, ok := w.notifiers[key]; ok {
		return fmt.Errorf("notifier %q is already registrered", key)
	}

	w.logger.Debugf("Added notifier %q to watcher", key)
	w.notifiers[key] = n
	return nil
}

// Run the jobs once.
func (w *Watcher) Run() {
	w.logger.Info("Running watcher ...")

	for _, job := range w.config.Jobs {
		if err := w.do(&job); err != nil {
			w.logger.Error(err)
		}
	}

	time.Sleep(1 * time.Second)

	w.logger.Info("done")
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

	var diff []api.Diff
	newValues := w.transformResults(pmJob.Results)
	for key, newVal := range newValues {
		oldVal, ok := oldValues[key]

		if job.NotifyOnChangeOnly && ok && oldVal != "" && oldVal == newVal {
			continue
		}

		diff = append(diff, api.Diff{
			Item:     key,
			OldValue: oldVal,
			NewValue: newVal,
		})
	}

	if len(diff) > 0 {
		err := w.notify(job, diff)
		if err != nil {
			return err
		}
	}

	w.logger.Infof("Done running job %s", job.Name)

	return w.setValues(job.Name, newValues)
}

func (w *Watcher) notify(job *api.Job, diff []api.Diff) error {
	for _, notify := range job.Notify {
		not, ok := w.notifiers[notify.Type]
		if !ok {
			return fmt.Errorf("notifier %q not found. It is either not available or not enabled", notify.Type)
		}

		if err := not.Notify(job.Name, notify.Value, diff); err != nil {
			return fmt.Errorf("failed to notify by %q: %v", notify.Type, err)
		}
	}

	return nil
}

// RegisterCronJobs registers all jobs taken from config at the given cron instance
func (w *Watcher) RegisterCronJobs(cron *cron.Cron) {
	for _, job := range w.config.Jobs {
		cron.AddFunc(job.Schedule, func() {

			w.do(&job)
		})
	}
}

func (w *Watcher) transformResults(values map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range values {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}