package uptimekuma

type Monitor struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	Hostname      string `json:"hostname"`
	Type          string `json:"type"`
	Interval      int    `json:"interval"`
	Active        bool   `json:"active"`
	Status        int    `json:"status"`
	UpsideDown    bool   `json:"upsideDown"`
	MaxRetries    int    `json:"maxretries"`
	RetryInterval int    `json:"retryInterval"`
	Description   string `json:"description"`
	Tags          []Tag  `json:"tags"`
}

type CreateMonitorInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Interval int    `json:"interval"`
}

type Heartbeat struct {
	ID        int    `json:"id"`
	MonitorID int    `json:"monitorID"`
	Status    int    `json:"status"`
	Time      string `json:"time"`
	Msg       string `json:"msg"`
	Ping      int    `json:"ping"`
	Duration  int    `json:"duration"`
}

type StatusPage struct {
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Published   bool   `json:"published"`
	Icon        string `json:"icon"`
	Theme       string `json:"theme"`
}

type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type HealthStatus struct {
	Status string `json:"status"`
}
