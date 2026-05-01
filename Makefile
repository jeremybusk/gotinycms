.PHONY: dev build package run web docker-up docker-build docker-down

web:
	cd web && npm install && npm run build

build:
	bash scripts/build.sh

package:
	bash scripts/package.sh

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
