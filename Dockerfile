FROM golang:1.18

WORKDIR /app

COPY . .

RUN go build -o practica-app

EXPOSE 8080

CMD ["./practica"]