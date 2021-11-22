
FROM golang:alpine as builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
ENV PORT=9000
WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o main .
WORKDIR /dist
RUN cp /build/main .
EXPOSE 9000
CMD ["/dist/main"]