FROM golang AS builder

WORKDIR /go/src/github.com/yousseffarkhani/playground/backend2
ADD . .

# Downloads all dependecies
RUN go get ./
RUN CGO_ENABLED=0 GOOS=linux go build -o playground

FROM alpine:latest AS production
COPY --from=builder /go/src/github.com/yousseffarkhani/playground/backend2 .
CMD ["./playground"]