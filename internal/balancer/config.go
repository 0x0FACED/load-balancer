package balancer

type Config struct {
	// Type can be "round_robin", "least_conn". Change it in config.json
	Type        BalancerType      `json:"type"`
	HealthCheck HealthCheckConfig `json:"healthcheck"`
	Backends    []string          `json:"backends"`
}

type HealthCheckConfig struct {
	Interval int `json:"interval"`
	Timeout  int `json:"timeout"`
}
