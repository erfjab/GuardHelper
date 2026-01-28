package users

type UsersResponse struct {
	Username            string                            `json:"username"`
	Status              string                            `json:"status"`
	Proxies             map[string]map[string]interface{} `json:"proxies,omitempty"`
	Inbounds            map[string][]string               `json:"inbounds,omitempty"`
	LifeTimeUsedTraffic int64                             `json:"life_time_used_traffic"`
	CreatedAt           string                            `json:"created_at"`
}