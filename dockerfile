FROM golang AS build

RUN mkdir /build
WORKDIR /build
COPY . /build

RUN go build /build/leaseplan-bot.go

FROM archlinux

WORKDIR /opt
COPY --from=build /build/leaseplan-bot /opt/leaseplan-bot

ENTRYPOINT ["/opt/leaseplan-bot"]