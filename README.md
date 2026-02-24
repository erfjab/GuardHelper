# GuardHelper

A simple API server built with Go and Fiber framework.

## Install

Download and install the correct build for your server:

```bash
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
  BINARY="guardhelper-linux-arm64"
else
  BINARY="guardhelper-linux-amd64"
fi
curl -L "https://github.com/erfjab/GuardHelper/releases/latest/download/${BINARY}" -o /usr/local/bin/guardhelper
chmod +x /usr/local/bin/guardhelper
sudo mkdir -p /etc/guardhelper
curl -L https://raw.githubusercontent.com/erfjab/guardhelper/refs/heads/master/.env.example -o /etc/guardhelper/.env
sudo nano /etc/guardhelper/.env
```

```bash
sudo bash -c 'cat > /etc/systemd/system/guardhelper.service << "EOF"
[Unit]
Description=GuardHelper API Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/guardhelper
ExecStart=/usr/local/bin/guardhelper
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable guardhelper
systemctl start guardhelper
```

## Logs

View live logs:

```bash
sudo journalctl -u guardhelper -f
```

View recent logs:

```bash
sudo journalctl -u guardhelper -n 100
```

## Update

Download the new version and restart:

```bash
sudo systemctl stop guardhelper
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
  BINARY="guardhelper-linux-arm64"
else
  BINARY="guardhelper-linux-amd64"
fi
curl -L "https://github.com/erfjab/GuardHelper/releases/latest/download/${BINARY}" -o /usr/local/bin/guardhelper
chmod +x /usr/local/bin/guardhelper
sudo systemctl start guardhelper
```

## Uninstall

Stop and remove the service:

```bash
sudo systemctl stop guardhelper
sudo systemctl disable guardhelper
sudo rm /etc/systemd/system/guardhelper.service
sudo rm /usr/local/bin/guardhelper
sudo rm -rf /etc/guardhelper
sudo systemctl daemon-reload
```
