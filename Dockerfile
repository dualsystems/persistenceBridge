FROM golang:1.16-alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /go/src/github.com/persistenceOne/persistenceBridge

# Add source files
COPY . .

RUN go version

# Install minimum necessary dependencies, build persistenceBridge, remove packages
RUN apk add --no-cache $PACKAGES && make install

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/app" \
    --shell "/sbin/nologin" \
    --uid "${UID}" \
    "${USER}"
USER 10001

WORKDIR /app

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/persistenceBridge /usr/bin/persistenceBridge

# Run persistenceBridge by default, omit entrypoint to ease using container with cli
CMD ["persistenceBridge"]
