package balancer

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type HealthChecker struct {
	servers []*Server
	client  *http.Client
	stop    chan struct{}
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// создает новый HealthChecker
func NewHealthChecker(servers []*Server) *HealthChecker {
	return &HealthChecker{
		servers: servers,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		stop: make(chan struct{}),
	}
}

// запускает периодические проверки здоровья
func (hc *HealthChecker) Start(interval time.Duration) {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stop:
				return
			}
		}
	}()
}

// останавливает проверки здоровья
func (hc *HealthChecker) Stop() {
	close(hc.stop)
	hc.wg.Wait()
}

// проверяет все серверы
func (hc *HealthChecker) checkAll() {
	var wg sync.WaitGroup

	hc.mu.RLock()
	servers := make([]*Server, len(hc.servers))
	copy(servers, hc.servers)
	hc.mu.RUnlock()

	for _, server := range servers {
		wg.Add(1)
		go func(s *Server) {
			defer wg.Done()
			hc.checkServer(s)
		}(server)
	}

	wg.Wait()
}

// проверяет здоровье одного сервера
func (hc *HealthChecker) checkServer(server *Server) {
	resp, err := hc.client.Get(server.URL.String() + "/health")
	if err != nil {
		hc.handleUnhealthy(server, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hc.handleUnhealthy(server, nil)
		return
	}

	hc.handleHealthy(server)
}

func (hc *HealthChecker) handleHealthy(server *Server) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if !server.Healthy {
		log.Printf("Server %s is now healthy", server.Config.Name)
		server.Healthy = true
	}
}

func (hc *HealthChecker) handleUnhealthy(server *Server, err error) {
	server.mu.Lock()
	defer server.mu.Unlock()

	if server.Healthy {
		if err != nil {
			log.Printf("Server %s is unhealthy: %v", server.Config.Name, err)
		} else {
			log.Printf("Server %s is unhealthy: bad status code", server.Config.Name)
		}
		server.Healthy = false
	}
}