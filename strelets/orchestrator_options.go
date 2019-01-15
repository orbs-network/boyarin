package strelets

type OrchestratorOptions struct {
	StorageDriver  string                 `json:"storage-driver"`
	StorageOptions map[string]interface{} `json:"storage-options"`
}
