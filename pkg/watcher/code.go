package watcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Scalify/puppet-master-client-go"
	"github.com/Scalify/website-content-watcher/pkg/api"
)

// loadFile fetches the content of a file as a string
// nolint: gosec
func (w *Watcher) loadFile(fileName string) (string, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to load file %q: %v", fileName, err)
	}

	return string(b), nil
}

// loadJSONFile loads the content of a file in the given target
// nolint: gosec
func (w *Watcher) loadJSONFile(fileName string, target interface{}) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to load file %q: %v", fileName, err)
	}

	if err := json.Unmarshal(b, target); err != nil {
		return fmt.Errorf("failed to decode json file %q content: %v", fileName, err)
	}

	return nil
}

// loadJob reads the job config files from disk and returns a jobRequest object,
// prepared for execution by the puppet master.
func (w *Watcher) loadJob(job *api.Job) (*puppetmaster.JobRequest, error) {
	codeFile := w.resolvePath(job.CodeFile)
	codeDir := filepath.Dir(codeFile)
	code, err := w.loadFile(codeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read code file %q: %v", codeFile, err)
	}

	varsFile := job.VarsFile
	vars := make(map[string]string)
	if varsFile == "" {
		varsFile = filepath.Join(codeDir, "vars.json")
	}
	varsFile = w.resolvePath(varsFile)
	if err = w.loadJSONFile(varsFile, &vars); err != nil {
		return nil, fmt.Errorf("failed to load vars file %q: %v", varsFile, err)
	}

	modulesDir := job.ModulesDir
	if modulesDir == "" {
		modulesDir = filepath.Join(codeDir, "modules")
	}
	modulesDir = w.resolvePath(modulesDir)
	modules, err := w.loadModules(modulesDir)
	if err != nil {
		return nil, err
	}

	w.logger.Debugf("Loaded code file %s, vars file %s, modules dir %s", codeFile, varsFile, modulesDir)

	jobReq := &puppetmaster.JobRequest{
		Code:    code,
		Vars:    vars,
		Modules: modules,
	}

	return jobReq, nil
}

// loadModules reads modules from disk, half way intelligently
func (w *Watcher) loadModules(modulesDir string) (map[string]string, error) {
	modules := make(map[string]string)

	files, err := ioutil.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return modules, err
	}

	var moduleFiles []string
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".mjs" {
			continue
		}

		moduleFiles = append(moduleFiles, filepath.Join(modulesDir, f.Name()))
	}

	for _, m := range moduleFiles {
		name := strings.Replace(filepath.Base(m), filepath.Ext(m), "", -1)
		modules[name], err = w.loadFile(m)
		if err != nil {
			return modules, err
		}
	}

	return modules, nil
}

func (w *Watcher) resolvePath(filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}

	return filepath.Join(filepath.Dir(w.configFile), filePath)
}
