FROM golang:1.13.5

RUN curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

WORKDIR /go/src/benchmark
COPY . .

RUN go get -d -v ./... github.com/pkg/errors
RUN go build -v -o benchmark ./...

ENTRYPOINT ["./benchmark"]