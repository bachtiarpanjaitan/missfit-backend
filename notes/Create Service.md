## Build App
go build -o missfit

## Create Service systemd
sudo gedit /etc/systemd/system/missfit.service

with content

```
[Unit]
Description=IhandLumos Backend
After=network.target

[Service]
Type=simple
User=bachtiar
WorkingDirectory=/home/bachtiar/srv/missfit-backend
ExecStart=/home/bachtiar/srv/missfit-backend/missfit
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target

```
## Reload Daemon
sudo systemctl daemon-reload
sudo systemctl start missfit

## Check status service
sudo systemctl status missfit

## Auto restart
sudo systemctl enable missfit



## When Update Code

go build -o missfit
sudo systemctl restart missfit
