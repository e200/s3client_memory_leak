# How to reproduce

```
docker-compose up -d
go run main.go
```

Wait for it to finish uploading the files, then:

```
curl http://localhost:8000/debug/pprof/allocs > allocs.mem
go tool pprof allocs.mem
```

Inside pprof run:

```
top10
```
