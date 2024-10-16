FROM golang:latest as builder

WORKDIR /app

COPY ./kademlia ./kademlia
COPY ./main ./main
COPY ./cli ./cli

RUN go mod init d7024e
RUN go mod tidy

WORKDIR /app/main
RUN go build -o /app/kademlia_main .

FROM golang:latest

RUN apt-get update && apt-get install -y \
    hping3 \
    netcat-openbsd

WORKDIR /app
COPY --from=builder /app/kademlia_main /app/kademlia_main

RUN mv /app/kademlia_main /usr/local/bin/kademlia

RUN chmod +x /usr/local/bin/kademlia

CMD ["kademlia"]
