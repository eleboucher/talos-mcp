FROM golang:1.26 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG GIT_COMMIT=unknown

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${GIT_COMMIT}" \
    -o talos-mcp ./cmd/talos-mcp

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/talos-mcp .
USER 65532:65532

ENTRYPOINT ["/talos-mcp"]
