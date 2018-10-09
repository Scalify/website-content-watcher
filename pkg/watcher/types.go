package watcher

import (
	"github.com/Scalify/puppet-master-client-go"
	"github.com/Scalify/website-content-watcher/pkg/api"
)

type storageClient interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Del(key string) error
}

type puppetMasterClient interface {
	CreateJob(jobRequest *puppetmaster.JobRequest) (*puppetmaster.Job, error)
	GetJob(uuid string) (*puppetmaster.Job, error)
	DeleteJob(uuid string) error
	ExecuteSync(jobRequest *puppetmaster.JobRequest) (*puppetmaster.Job, error)
}

type notifier interface {
	Key() string
	Notify(jobName, target string, diff []api.Diff, newValues map[string]string) error
}
