
all: build

prepare:
	npm install

build:
	npm run build
	cd server && go build

serve: build
	./server/ace_away -path dist
