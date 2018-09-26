package watcher

import (
	"github.com/Scalify/puppet-master-client-go"
	"github.com/Scalify/website-content-watcher/pkg/api"
)

func (w *Watcher) executeJob(job *api.Job) (*puppetmaster.Job, error) {
	//b, _ := ioutil.ReadFile("job.json")
	//pmJob := &puppetmaster.Job{}
	//json.Unmarshal(b, pmJob)
	//return pmJob, nil

	pmJobReq, err := w.loadJob(job)
	if err != nil {
		return nil, err
	}
	pmJob, err := w.puppet.ExecuteSync(pmJobReq)
	if err != nil {
		return nil, err
	}

	return pmJob, nil
}
