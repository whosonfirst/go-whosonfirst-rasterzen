tools:
	go build -mod vendor -o bin/wof-rasterzen-seed cmd/wof-rasterzen-seed/main.go

docker:
	docker build -t wof-rasterzen-seed .
