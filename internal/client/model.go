package client

type Client struct {
	ID         string `json:"id"`
	Capacity   int    `json:"capacity"`
	RefillRate int    `json:"refill_rate"`
}
