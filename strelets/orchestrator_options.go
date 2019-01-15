package strelets

type OrchestratorOptions struct {
	StorageDriverValue  string                 `json:"storage-driver"`
	StorageOptionsValue map[string]interface{} `json:"storage-options"`
}

func (o *OrchestratorOptions) StorageDriver() string {
	return o.StorageDriverValue
}

func (o *OrchestratorOptions) StorageOptions() map[string]interface{} {
	return o.StorageOptionsValue
}
