package gcp

type Filter struct {
	TagRegexString  string
	NameRegexString string
	Status          string
}

const (
	InstanceAll        = "ALL"
	InstanceRunning    = "RUNNING"
	InstanceTerminated = "TERMINATED"
	ImageReady         = "READY"
	ImageFailed        = "FAILED"
)
