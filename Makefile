.PHONY: serve build photos

build:
	cd build && go run .

serve:
	uv run serve.py

photos:
	./scripts/optimize-photos.sh photos/2025-sicily --html
	./scripts/optimize-photos.sh photos/2025-slovenia --html
