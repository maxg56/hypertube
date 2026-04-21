package services

// ServiceConfig represents the configuration for an upstream service
type ServiceConfig struct {
	Name    string
	URL     string
}

var services map[string]ServiceConfig

// InitServices initializes the service configuration map
func InitServices() {
	services = map[string]ServiceConfig{
		"auth": {
			Name: "auth-service",
			URL:  "http://auth-service:8001",
		},
		"user": {
			Name: "user-service",
			URL:  "http://user-service:8002",
		},
		"library": {
			Name: "library-service",
			URL:  "http://library-service:8003",
		},
		"torrent": {
			Name: "torrent-service",
			URL:  "http://torrent-service:8004",
		},
		"comment": {
			Name: "comment-service",
			URL:  "http://comment-service:8005",
		},
		"worker": {
			Name: "worker-service",
			URL:  "http://worker-service:8006",
		},
	}
}

// GetService returns service configuration by name
func GetService(name string) (ServiceConfig, bool) {
	service, exists := services[name]
	return service, exists
}

// GetServicesStatus returns the status map of all configured services
func GetServicesStatus() map[string]string {
	status := make(map[string]string)
	for name, service := range services {
		status[name] = service.URL
	}
	return status
}
