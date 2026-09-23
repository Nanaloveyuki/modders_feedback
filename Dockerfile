FROM public.ecr.aws/docker/library/node:22-alpine AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM public.ecr.aws/docker/library/golang:1.24-alpine AS server-build
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/feedback ./cmd/feedback

FROM public.ecr.aws/docker/library/alpine:3.22
RUN addgroup -S -g 65532 nonroot && adduser -S -D -H -u 65532 -G nonroot nonroot \
    && mkdir -p /data && chown nonroot:nonroot /data
WORKDIR /app
COPY --from=server-build --chown=nonroot:nonroot /out/feedback /app/feedback
COPY --from=web-build --chown=nonroot:nonroot /src/web/dist /app/web
USER nonroot
ENV LISTEN_ADDR=:8080 DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/feedback"]
