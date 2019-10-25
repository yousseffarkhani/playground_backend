# FROM golang AS builder
FROM golang

WORKDIR /go/src/github.com/yousseffarkhani/playground/backend2
ADD . .

# Downloads all dependecies
RUN go get ./
# Single staged
RUN go install

CMD backend2

# Multi staged not working properly
# RUN CGO_ENABLED=0 GOOS=linux go build -o playground

# FROM alpine:latest AS production
# COPY --from=builder /go/src/github.com/yousseffarkhani/playground/backend2 .
# CMD ["./playground"]