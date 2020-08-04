# SPDX-FileCopyrightText: 2020 jecoz
#
# SPDX-License-Identifier: BSD-3-Clause

FROM golang:1.14.6-buster AS builder

ENV GO111MODULE=on CGO_ENABLED=0

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy code into /build
COPY . .

RUN go build -o flexi cmd/flexi/main.go

WORKDIR /dist
RUN cp /build/flexi .

FROM jecoz/9image
WORKDIR /

RUN apt-get update \
    && apt-get install -y ca-certificates fuse \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /dist/flexi /usr/bin

EXPOSE 564

# Command to run
ENTRYPOINT ["flexi", "-port", "564", "-m", "mnt"]
