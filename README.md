# goprof-optimizer

**Production-grade Go runtime profiler + memory retention analyzer.**  
Built with enterprise patterns, Prometheus metrics, pprof integration, structured logging, and configurable sampling.

---

## ✨ Features

- Runtime memory sampling (`runtime.MemStats`)
- Allocation tracking (`TrackAllocation`)
- Retention estimation per type/tag
- Heuristic optimization suggestions
- Prometheus metrics (`/metrics`)
- Health endpoints:
  - `/health/live`
  - `/health/ready`
- REST API:
  - `/v1/metrics/latest`
  - `/v1/metrics/history?limit=N`
  - `/v1/suggestions`
  - `/v1/alerts`
- pprof endpoints
- Fully configurable via YAML or environment

---

## 🚀 Running locally

```bash
make run
````

Or manually:

```bash
go run ./cmd/profiler -config=config.example.yaml
```

---

## 🐳 Docker

```bash
docker build -t goprof-optimizer .
docker run -p 8080:8080 goprof-optimizer
```

---

## 📊 Prometheus

If you're using `docker-compose.yml`, a Prometheus instance is included.

Open:
[http://localhost:9090](http://localhost:9090)

---

## 🧪 Testing

```bash
make test
```

---

## 📂 Project Structure

(omitted here – see discussion and directories)

---

## 📝 License

MIT (or whatever you want)

```
