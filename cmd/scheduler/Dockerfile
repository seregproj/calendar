# Собираем в гошке
FROM golang:1.17 as build

ENV BIN_FILE /opt/scheduler/scheduler-app
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
        -o ${BIN_FILE} cmd/scheduler/*.go

# На выходе тонкий образ
FROM alpine:3.9

LABEL SERVICE="scheduler"
LABEL MAINTAINERS="kozirevss@gmail.com"

ENV BIN_FILE "/opt/scheduler/scheduler-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/scheduler/scheduler_config.yml
COPY ./configs/scheduler_config.yml ${CONFIG_FILE}

RUN mkdir -p /var/log

CMD LOG_FILE=/var/log/scheduler.log ${BIN_FILE} -config ${CONFIG_FILE}
