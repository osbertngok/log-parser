#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${DIR}/models
go generate
cd ${DIR}
go build ${DIR}/cmd/log2csv
docker build -t osbertngok/log-parser:master
