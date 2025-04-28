package balancer

type Config struct {
	Backends []string `json:"backends"`
}
