.PHONY: serve photos

serve:
	python3 serve.py

photos:
	./scripts/optimize-photos.sh photos/2025-sicily --html
	./scripts/optimize-photos.sh photos/2025-slovenia --html
