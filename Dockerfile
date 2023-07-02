FROM golang:alpine as builder

RUN apk add nodejs npm --no-cache

WORKDIR /app

COPY . /app

RUN go mod download
RUN npm install
RUN ./build.sh

FROM scratch

COPY --from=builder /app/build/neohealth .

EXPOSE 1234
CMD ["./neohealth"]
