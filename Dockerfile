ARG GO_VERSION=1.15
ARG FROM_IMAGE=alpine:latest

FROM golang:${GO_VERSION}-alpine3.12 AS builder
RUN apk add --update --no-cache bash ca-certificates make curl git mercurial tzdata

# Install kubectl
ARG KUBECTL_VERSION=v1.16.1
RUN curl -L -s https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl
RUN curl -L -s https://amazon-eks.s3-us-west-2.amazonaws.com/1.14.6/2019-08-22/bin/linux/amd64/aws-iam-authenticator -o /usr/local/bin/aws-iam-authenticator

RUN chmod +x /usr/local/bin/*

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

ARG VERSION

COPY . /build
RUN make build-release

FROM ${FROM_IMAGE}
RUN apk add --update --no-cache openssh-client
COPY --from=builder /usr/local/bin/* /bin/
COPY --from=builder /build/build/banzai /bin/

ENTRYPOINT [ "/bin/banzai" ]
