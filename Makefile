all: compile

compile:
	go build -o benchmark .

docker-compile:
	CGO_ENABLED=0 GOOS=linux go build -o benchmark .

run: compile
	./benchmark

docker-build:
	docker build -t antidote-benchmark .

docker-run: docker-build
	./script/run.sh