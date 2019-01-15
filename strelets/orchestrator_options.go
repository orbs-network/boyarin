package strelets

type OrchestratorOptions struct {
	StorageDriverValue  string            `json:"storage-driver"`
	StorageOptionsValue map[string]string `json:"storage-options"`
}

func (o *OrchestratorOptions) StorageDriver() string {
	return o.StorageDriverValue
}

func (o *OrchestratorOptions) StorageOptions() map[string]string {
	return o.StorageOptionsValue
}
