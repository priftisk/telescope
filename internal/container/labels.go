package container

const (
	ProxyHost = "proxy.host"
	ProxyPort = "proxy.port"
	ProxyPath = "proxy.path"
)

type Labels struct {
	ProxyHost string `json:"proxy_host"`
	ProxyPort string `json:"proxy_port"`
	ProxyPath string `json:"proxy_path"`
}

func (labels Labels) IsValid() bool {
	return labels.ProxyHost != "" && labels.ProxyPort != ""
}
