FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api && CGO_ENABLED=0 go build -o /relayer ./cmd/relayer
# Created here so the uploads volume starts out owned by the unprivileged user.
RUN mkdir -p /data/uploads

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /api /api
COPY --from=build /relayer /relayer
COPY --from=build --chown=nonroot:nonroot /data /data
EXPOSE 8080
ENTRYPOINT ["/api"]
