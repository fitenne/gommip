FROM alpine:latest

COPY gommip config.yml /app/
RUN apk add curl

CMD ["/app/gommip", "-config", "/app/config.yml"]

EXPOSE 8080/tcp

HEALTHCHECK --interval=5m --timeout=3s --start-period=3m \
  CMD curl -f http://localhost:8080/health || exit 1