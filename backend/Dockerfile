# --- build ---
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/kanban .

# --- runtime ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/kanban /app/kanban
RUN mkdir -p /app/data && chown -R app:app /app
USER app

ENV PORT=8080 \
    DATA_FILE=/app/data/tasks.json \
    CORS_ORIGINS=*

EXPOSE 8080
CMD ["/app/kanban"]
