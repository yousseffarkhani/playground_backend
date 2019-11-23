FROM golang AS builder

WORKDIR /go/src/github.com/yousseffarkhani/playground/backend2
ADD . .

# Downloads all dependecies
RUN go get ./
RUN CGO_ENABLED=0 GOOS=linux go build -o backend2

FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/yousseffarkhani/playground/backend2 .

CMD ["./backend2"]