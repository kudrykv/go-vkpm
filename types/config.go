package types

type Config struct {
	Domain         string  `yaml:"domain"`
	DefaultProject string  `yaml:"default_project"`
	Cookies        Cookies `yaml:"cookies"`
}

type Cookies struct {
	CSRFToken string `yaml:"csrftoken"`
	SessionID string `yaml:"sessionid"`
}
