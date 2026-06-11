FROM alpine:3.21

RUN apk add --no-cache nftables iproute2

WORKDIR /app
COPY cmd/simulator/entrypoint.sh .
COPY simulator-linux /app/server-simulator

RUN chmod +x entrypoint.sh server-simulator

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
