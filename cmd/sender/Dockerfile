# Собираем в гошке
FROM golang:1.17 as build

ENV BIN_FILE /opt/sender/sender-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -o ${BIN_FILE} cmd/sender/*.go

# На выходе тонкий образ
FROM alpine:3.9

LABEL SERVICE="sender"
LABEL MAINTAINERS="kozirevss@gmail.com"

ENV BIN_FILE "/opt/sender/sender-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/sender/sender_config.yml
COPY ./configs/sender_config.yml ${CONFIG_FILE}

RUN mkdir -p /var/log

CMD LOG_FILE=/var/log/sender.log ${BIN_FILE} -config ${CONFIG_FILE}
