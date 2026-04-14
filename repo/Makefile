.PHONY: test build docker docker-down clean

test:
	go test ./... -coverprofile=cov.out -cover -v

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

docker:
	docker compose up --build -d

docker-down:
	docker compose down -v

clean:
	rm -rf bin/ cov.out data/
