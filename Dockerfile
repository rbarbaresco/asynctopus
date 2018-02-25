FROM golang:1.8

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY src /go/src/app

RUN go get -d -v ./...
RUN go install -v ./...
RUN yes | go build -o asynctopus

COPY . /usr/src/app

EXPOSE 8079

CMD ["./asynctopus"]
