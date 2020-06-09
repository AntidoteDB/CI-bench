FROM golang:1.13.5

RUN curl -L "https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

RUN apt-get update && apt-get install -y python3 python3-pip && rm -rf /var/lib/apt/lists/* && pip3 install numpy pandas seaborn

WORKDIR /go/src/benchmark
COPY . .

RUN go mod init
RUN go get -d -v ./... 
RUN go get -d -v github.com/pkg/errors
RUN go get -d -v github.com/docker/docker@master
RUN go build -v -o benchmark ./...

ENTRYPOINT ["./benchmark"]
