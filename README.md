# LoadBalancerGo

Simple learning load-balancer in Go with:

- Round-robin and Least-Connections algorithms
- Passive and active health checks
- Retries and per-request timeouts

Structure
- `pkg/loadbalancer` — core library (Backends, ServerPool, LoadBalancer handler)
- `cmd/loadbalancer` — load balancer CLI (entrypoint)
- `cmd/backend-server` — tiny test backend used for local testing

Quick Start

Run three test backends (each in its own terminal):
```bash
cd /home/yanazyab/02Personal/16LoadBalancerGo
go run ./cmd/backend-server -port=8081 &
go run ./cmd/backend-server -port=8082 &
go run ./cmd/backend-server -port=8083 -sleep=1000 &
```

Run the load balancer (round-robin):
```bash
go run ./cmd/loadbalancer -port=8000 -backends=http://localhost:8081,http://localhost:8082,http://localhost:8083 -mode=rr
```

Run the load balancer (least-connections):
```bash
go run ./cmd/loadbalancer -port=8001 -backends=http://localhost:8081,http://localhost:8082,http://localhost:8083 -mode=least
```

Notes
- Health checks run every 20s by default.
- Adjust timeouts and retry limits in `pkg/loadbalancer/handler.go`.

