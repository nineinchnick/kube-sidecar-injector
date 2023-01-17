# Build the sidecar-injector binary
FROM golang:1.19 as builder

WORKDIR /

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY ./ src/

RUN cd src \
    && CGO_ENABLED=0 go build -a -o sidecar-injector .


FROM alpine:latest

# install curl for prestop script
RUN apk --no-cache add curl

WORKDIR /

COPY --from=builder /src/sidecar-injector .
ADD ./prestop.sh .

USER 65532:65532

ENTRYPOINT ["./sidecar-injector"]
