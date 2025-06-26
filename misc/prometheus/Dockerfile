FROM ubuntu:24.10

ENV LANG C.UTF-8

RUN apt-get update && apt-get install -y ca-certificates && apt-get clean

WORKDIR /opt/

COPY alertmanager_tg_bot /opt/

CMD ["/opt/alertmanager_tg_bot"]
