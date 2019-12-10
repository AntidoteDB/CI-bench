all: compile

compile:
	go build -o Benchmark .

run: compile
	./Benchmark