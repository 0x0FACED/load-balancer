package limiter

type Config struct {
	Capacity        int `json:"capacity"`
	Rate            int `json:"rate"`
	RefillIntrerval int `json:"refill_interval"`
	TTL             int `json:"ttl"`
}
