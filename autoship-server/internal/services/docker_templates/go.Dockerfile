FROM golang:1.20
WORKDIR /app
COPY . .
RUN go build -o app .
EXPOSE 8080
CMD ["./app"]
