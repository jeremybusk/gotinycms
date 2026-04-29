.PHONY: dev build run web

web:
	cd web && npm install && npm run build

build: web
	go build -trimpath -ldflags='-s -w' -o bin/tinycms ./cmd/tinycms

run: build
	./bin/tinycms

dev:
	mkdir -p data/uploads
	( cd web && npm install && npm run dev ) & go run ./cmd/tinycms
