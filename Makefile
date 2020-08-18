clean:
	@rm -rf mc/
	@rm -rf build/
	@echo "[OK] All files cleaned!"

run:
	@go run ./...

build:
	@go build -o build/mcserverlauncher
	@echo "[OK] Application build"

test:
	@go test ./...
	@echo "[OK] All tests run!"