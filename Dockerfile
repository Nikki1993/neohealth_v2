FROM golang:alpine

RUN apk add nodejs npm --no-cache

WORKDIR /app

COPY . /app

RUN go mod download
RUN npm install
RUN ./build.sh

EXPOSE 1234
CMD ["./build/neohealth"]
