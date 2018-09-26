package api

// Config represents a configuration file
type Config struct {
	Jobs []Job `json:"jobs"`
}

// Job entry of a config file. Defines what to execute when.
type Job struct {
	Name               string        `json:"name"`
	Schedule           string        `json:"schedule"`
	Notify             []NotifyEntry `json:"notify"`
	NotifyOnChangeOnly bool          `json:"notify_on_change_only"`
	CodeFile           string        `json:"code_file"`
	VarsFile           string        `json:"vars_file"`
	ModulesDir         string        `json:"modules_dir"`
}

// NotifyEntry defines whom to notify on change
type NotifyEntry struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Diff defines the diff between two watch states over time
type Diff struct {
	Item, OldValue, NewValue string
}
