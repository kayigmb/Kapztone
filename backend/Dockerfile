# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
# -trimpath and -s -w drop paths and debug symbols: smaller binaries, same code.
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /relayer ./cmd/relayer
# Created here so the uploads volume starts out owned by the unprivileged user.
RUN mkdir -p /data/uploads

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /api /api
COPY --from=build /relayer /relayer
COPY --from=build --chown=nonroot:nonroot /data /data
EXPOSE 8080
ENTRYPOINT ["/api"]
