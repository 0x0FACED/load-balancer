package balancer

type Config struct {
	// Type can be "round_robin", "least_conn". Change it in config.json
	Type     BalancerType `json:"type"`
	Backends []string     `json:"backends"`
}
