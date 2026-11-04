BIN := bin/leads

.PHONY: build test dev dev-api dev-web web-build import clean

web/node_modules: web/package.json web/package-lock.json
	cd web && npm ci
	@touch $@

## web-build: type-check and build the Vue app into web/dist
web-build: web/node_modules
	cd web && npm run build

## build: build the Vue app, then compile everything into one binary at bin/leads
build: web-build
	CGO_ENABLED=0 go build -o $(BIN) .

## test: run the Go tests
test: web-build
	go test ./...

## import: build, then import leads.csv into the database
import: build
	./$(BIN) import leads.csv

## dev: run the Go API (:8080) and the Vite dev server (:5173) together
dev: web-build
	$(MAKE) -j2 dev-api dev-web

dev-api:
	go run . serve

dev-web: web/node_modules
	cd web && npm run dev

clean:
	rm -rf bin web/dist
