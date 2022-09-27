FROM golang:1.17-alpine

ENV GO111MODULE=on
ENV GIN_MODE=release
ENV CGO_ENABLED=0

WORKDIR /app/cachetoonqueue

COPY . .

RUN go mod download
RUN go mod verify
RUN go build -o cache-toon-queue ./

ENTRYPOINT [ "./cache-toon-queue" ]