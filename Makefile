.PHONY: dev build run web docker-up docker-build docker-down

web:
	cd web && npm install && npm run build

build: web
	go build -trimpath -ldflags='-s -w' -o bin/uvoocms ./cmd/uvoocms

run: build
	./bin/uvoocms

dev:
	mkdir -p data/uploads
	( cd web && npm install && npm run dev ) & go run ./cmd/uvoocms

docker-build:
	docker compose build uvoocms

docker-up:
	docker compose up -d --build --remove-orphans uvoocms

docker-down:
	docker compose down
