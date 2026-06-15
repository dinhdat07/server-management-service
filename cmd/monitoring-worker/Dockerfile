FROM alpine:3.21

RUN apk add --no-cache iproute2

WORKDIR /app
COPY bin/monitoring-worker-linux /app/monitoring-worker
COPY cmd/monitoring-worker/docker-entrypoint.sh /app/docker-entrypoint.sh

RUN sed -i 's/\r$//' docker-entrypoint.sh && chmod +x docker-entrypoint.sh monitoring-worker

ENTRYPOINT ["./docker-entrypoint.sh"]
