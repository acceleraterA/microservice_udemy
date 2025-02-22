# base go image

FROM alpine:latest

RUN mkdir /app

COPY mailerServiceApp /app

CMD ["/app/mailerServiceApp"]