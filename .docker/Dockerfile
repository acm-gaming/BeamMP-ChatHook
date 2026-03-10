FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
COPY chathook-daemon ./chathook-daemon

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -trimpath -ldflags='-s -w' -o /out/chathook-daemon ./chathook-daemon/cmd/chathook-daemon

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
ENV UDP_PORT=30813
COPY --from=build /out/chathook-daemon /usr/local/bin/chathook-daemon
EXPOSE 30813/udp
ENTRYPOINT ["/usr/local/bin/chathook-daemon"]
