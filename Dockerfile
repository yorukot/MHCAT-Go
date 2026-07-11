FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
RUN set -eux; \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/mhcat-bot ./cmd/mhcat-bot; \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mhcat-shard-supervisor ./cmd/mhcat-shard-supervisor; \
    for command in mhcat-command-sync mhcat-mongo-audit mhcat-mongo-index mhcat-staging-preflight mhcat-economy-reset mhcat-work-payout mhcat-scheduler-lease; do \
      CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "/out/${command}" "./cmd/${command}"; \
    done

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 mhcat \
    && adduser -S -D -H -u 10001 -G mhcat mhcat

WORKDIR /app
COPY --from=build /out/ /usr/local/bin/
COPY --from=build --chown=mhcat:mhcat /src/asset/ ./asset/
COPY --from=build --chown=mhcat:mhcat /src/fonts/ ./fonts/

USER mhcat
ENV MHCAT_BOT_PATH=/usr/local/bin/mhcat-bot
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=30s --timeout=5s --start-period=120s --retries=3 \
  CMD expected="${MHCAT_DISCORD_SHARD_COUNT:-1}"; set -- $(pidof mhcat-bot 2>/dev/null || true); test "$#" -eq "$expected"

ENTRYPOINT ["/usr/local/bin/mhcat-shard-supervisor"]
