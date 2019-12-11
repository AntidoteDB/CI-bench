
## Build Steps

1. Install dependencies:

	go get github.com/pkg/errors
	rm -rf ~/go/src/github.com/docker/docker/vendor/github.com/docker/go-connections
	go get

2. Compile project:

	make
