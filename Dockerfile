FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 as builder

ARG TARGETOS TARGETARCH

RUN apk add --no-cache make

WORKDIR /tofulint
COPY . /tofulint
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build

FROM alpine:3.19

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tofulint/dist/tofulint /usr/local/bin

ENTRYPOINT ["tofulint"]
WORKDIR /data
