# this script will start
go run main.go &&
  wrk -t 32 -c 200 -d 200s --latency 'http://localhost:8080/single_flight' &&
  go tool pprof -http=':8000' 'http://localhost:8080/debug/pprof/profile\?debug\=1'
