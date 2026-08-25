# syntax=docker/dockerfile:1
ARG TF_VERSION=1.15.9
ARG GO_VERSION=1.25.8

FROM hashicorp/terraform:${TF_VERSION} AS terraform
FROM ghcr.io/opentofu/opentofu:${TF_VERSION}-minimal AS tofu

FROM golang:${GO_VERSION}-alpine AS base
RUN apk add --no-cache ca-certificates curl unzip git jq bash gcompat make
WORKDIR /workspace

COPY run-tests.sh /usr/local/bin/run-tests.sh
RUN chmod +x /usr/local/bin/run-tests.sh

FROM base AS with-terraform
COPY --from=terraform /bin/terraform /usr/local/bin/terraform
ENV TF_ACC_TERRAFORM_PATH=/usr/local/bin/terraform
ENTRYPOINT ["/usr/local/bin/run-tests.sh"]

FROM base AS with-opentofu
COPY --from=tofu /usr/local/bin/tofu /usr/local/bin/tofu
ENV TF_ACC_TERRAFORM_PATH=/usr/local/bin/tofu
ENTRYPOINT ["/usr/local/bin/run-tests.sh"]
