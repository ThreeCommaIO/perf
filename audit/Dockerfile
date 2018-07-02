FROM golang:latest as builder
COPY . /go/src/app
WORKDIR /go/src/app

RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app

FROM alpine:latest
# Copy our static executable
COPY --from=builder /go/src/app/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]