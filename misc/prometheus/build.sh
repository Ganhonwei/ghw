CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath --ldflags="-s -w" -o alertmanager_tg_bot ./

# docker build -t alertmanager_tg_bot:1 .

# node_exporter 主机指标监控
# /etc/systemd/system/node_exporter.service
# [Unit]
# Description=node_exporter
# After=network.target
# 
# [Service]
# Restart=on-failure
# ExecStart=/home/ubuntu/game/prometheus/node_exporter/node_exporter --web.listen-address=:9100
# 
# [Install]
# WantedBy=multi-user.target
