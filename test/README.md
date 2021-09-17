# Testing


## Profiling

go tool pprof  http://localhost:6060/debug/pprof/heap


k6 run k6/loadtest.js --vus 2 --duration 30s  --insecure-skip-tls-verify
