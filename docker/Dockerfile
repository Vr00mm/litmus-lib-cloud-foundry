# syntax=docker/dockerfile:1.3

FROM FROM golang:1.18.3 AS builder

WORKDIR /litmus-go

COPY ./src/go.mod ./src/go.sum /litmus-go/
RUN go mod download

ADD ./src /litmus-go
RUN CGO_ENABLED=0 go build -o /output/experiments ./bin/experiment

FROM alpine:3.15.0

LABEL maintainer="LitmusChaos"

# make a pipe fail on the first failure
SHELL ["/bin/sh", "-o", "pipefail", "-c"]

RUN apk --update add curl bash python3-dev py3-pip
RUN pip install --no-cache-dir setuptools chaostoolkit
RUN pip install -U chaostoolkit-cloud-foundry

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin

ADD ./cloud-foundry-experiments/ /litmus
COPY --from=builder /output/ /litmus
WORKDIR /litmus
