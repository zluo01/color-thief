FROM golang:alpine AS builder

WORKDIR /go/src/colorThief

COPY . .

# git
RUN apk update && apk add --no-cache git \
 && go get -d -v ./... \
 && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/main .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy our static executable.
COPY --from=builder /go/bin/main .

EXPOSE 8080

CMD ["./main"]
