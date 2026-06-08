# Build a static ccview binary, then ship it on a minimal base.
# The SQLite driver is pure Go (modernc.org/sqlite), so CGO stays off and the
# binary runs on distroless/static with no libc.
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=docker
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o /ccview ./cmd/ccview

FROM gcr.io/distroless/static-debian12
COPY --from=build /ccview /ccview
# CLAUDE_CONFIG_DIR points at the mounted data dir; ccview expects projects/
# (read) and ccview/ (its own DB + trash, read-write) underneath it.
ENV CLAUDE_CONFIG_DIR=/data
EXPOSE 12100
ENTRYPOINT ["/ccview", "--no-browser", "--bind", "0.0.0.0", "--port", "12100"]
