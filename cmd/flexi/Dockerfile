# SPDX-FileCopyrightText: 2020 jecoz
#
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.14.6-buster AS builder

ENV GO111MODULE=on CGO_ENABLED=0

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main cmd/flexi/main.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main .

# Build a small image
FROM jecoz/9image

RUN apt-get update \
    && apt-get install -y ca-certificates fuse \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /dist/main /

EXPOSE 564

# Command to run
ENTRYPOINT ["/main", "-port", "564"]
