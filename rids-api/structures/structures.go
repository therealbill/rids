package structures

type InstanceDataJSON struct {
	Config map[string]string `json:"config"`
	Info   string            `json:"info"`
	Error  string            `json:"pull_error"`
}
