FROM golang:1.10.2-alpine3.7

WORKDIR /app
COPY ./log2csv /app/log2csv
