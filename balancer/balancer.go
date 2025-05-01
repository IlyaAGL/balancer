package balancer

import (
	"container/heap"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/agl/balancer/config"
)

type Server struct {
    Config      config.ServerConfig
    URL         *url.URL
    Connections int
    Healthy     bool
    mu          sync.Mutex
    proxy       *httputil.ReverseProxy
}

type ServerHeap []*Server

func (h ServerHeap) Len() int           { return len(h) }
func (h ServerHeap) Less(i, j int) bool { 
    return float64(h[i].Connections)/float64(h[i].Config.Weight) < 
           float64(h[j].Connections)/float64(h[j].Config.Weight) 
}
func (h ServerHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *ServerHeap) Push(x interface{}) {
    *h = append(*h, x.(*Server))
}

func (h *ServerHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

type LoadBalancer struct {
	servers      []*Server
	heap         *ServerHeap
	healthChecker *HealthChecker
	mu           sync.RWMutex
}

func NewLoadBalancer(cfg config.Config) (*LoadBalancer, error) {
	lb := &LoadBalancer{
		heap: &ServerHeap{},
	}

	for _, srvCfg := range cfg.BackendServers {
		u, err := url.Parse(srvCfg.URL)
		if err != nil {
			return nil, err
		}

		server := &Server{
			Config:      srvCfg,
			URL:         u,
			Connections: 0,
			Healthy:     true,
		}

		server.proxy = httputil.NewSingleHostReverseProxy(u)
		lb.servers = append(lb.servers, server)
		heap.Push(lb.heap, server)
	}

	lb.healthChecker = NewHealthChecker(lb.servers)
	lb.healthChecker.Start(cfg.HealthCheck)

	return lb, nil
}

func (lb *LoadBalancer) Stop() {
	lb.healthChecker.Stop()
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    var selectedServer *Server
    for lb.heap.Len() > 0 {
        server := heap.Pop(lb.heap).(*Server)
        
        server.mu.Lock()
        if server.Healthy {
            server.Connections++
            selectedServer = server
            server.mu.Unlock()
            break
        }
        server.mu.Unlock()
    }

    if selectedServer == nil {
        http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
        return
    }

    selectedServer.proxy.ServeHTTP(w, r)
    heap.Push(lb.heap, selectedServer)
}