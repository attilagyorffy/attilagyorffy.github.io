.PHONY: serve build photos

build:
	cd build && go run .

serve:
	cd build && go run . serve

photos:
	cd build && go run . photos 2025-lipari
	cd build && go run . photos 2025-bled
