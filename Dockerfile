FROM golang:1.17

ENV X_CLIENT_KEY="lana-abugauch"

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["/app/cmd/api"]