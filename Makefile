.PHONY: dev build run web docker-up docker-build docker-down

web:
	cd web && npm install && npm run build

build: web
	go build -trimpath -ldflags='-s -w' -o bin/uvoominicms ./cmd/uvoominicms

run: build
	./bin/uvoominicms

dev:
	mkdir -p data/uploads
	( cd web && npm install && npm run dev ) & go run ./cmd/uvoominicms

docker-build:
	docker compose build uvoominicms

docker-up:
	docker compose up -d --build --remove-orphans uvoominicms

docker-down:
	docker compose down
