## Build App
go build -o lumos

## Create Service systemd
sudo gedit /etc/systemd/system/lumos.service

with content

```
[Unit]
Description=IhandLumos Backend
After=network.target

[Service]
Type=simple
User=bachtiar
WorkingDirectory=/home/bachtiar/srv/lumos
ExecStart=/home/bachtiar/srv/lumos/lumos
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target

```
## Reload Daemon
sudo systemctl daemon-reload
sudo systemctl start lumos

## Check status service
sudo systemctl status lumos

## Auto restart
sudo systemctl enable lumos



## When Update Code

go build -o lumos
sudo systemctl restart lumos
