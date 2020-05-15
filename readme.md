
## Build Steps

1. Clone this repository into your Go path.
2. Install dependencies:

		go get github.com/pkg/errors
		rm -rf ~/go/src/github.com/docker/docker/vendor/github.com/docker/go-connections
		go get

3. Compile project:

		make


## Running Benchmarks via Docker-compose


1. Build Benchmark docker image:

		docker build --no-cache -t antidote-benchmark .

2. Build the Antidote docker image. In the [Antidote repository](https://github.com/AntidoteDB/antidote) run:

		make docker-build

3. Download the Cadvisor image:

		docker pull google/cadvisor

4. Run the benchmark:

		cd script
		./run.sh -n default -c 4 -r 50000

	Reports can be found in the `output` folder.

## Options

For a list of configuration options run the executable with `--help`.