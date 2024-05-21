FROM ubuntu:latest

WORKDIR /app

COPY finaltask ./

COPY web ./web

CMD ["/app/finaltask"]