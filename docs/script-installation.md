Perfect 👌 Let’s turn this into a **step-by-step installation guide** for a fresh server setup.
Here’s a full structured document you can drop into your repo as `autoship-setup.md`:

---

# 🚀 AutoShip Deployment Handler — Setup Guide

This guide explains how to install, configure, and run the **AutoShip Deployment Handler** on a fresh Linux server (Fedora/Ubuntu/Debian).

---

## 1️⃣ Prerequisites

* A Linux server (Ubuntu/Debian/Fedora tested)
* Root or `sudo` access
* Python 3.8+ installed
* NGINX installed
* Certbot installed (`apt install certbot python3-certbot-nginx` or `dnf install certbot`)
* Cloudflare API token for DNS automation

---

## 2️⃣ Clone the Code

```bash
# Move to /opt and clone repo
cd /opt
sudo git clone https://github.com/your-org/autoship.git
sudo chown -R $USER:$USER /opt/autoship
```

Your main entrypoint should be:

```
/opt/autoship/main.py
```

---

## 3️⃣ Prepare Folders

```bash
# Webroot for ACME challenges
sudo mkdir -p /var/www/autoship

# Certbot dirs
sudo mkdir -p /var/lib/autoship/certbot/config
sudo mkdir -p /var/lib/autoship/certbot/work
sudo mkdir -p /var/lib/autoship/certbot/logs

# Nginx folders (if not already existing)
sudo mkdir -p /etc/nginx/sites-available
sudo mkdir -p /etc/nginx/sites-enabled
```

---

## 4️⃣ Systemd Service Setup
copy in /opt/autoship/
```bash
sudo cp -r . /opt/autoship/
```

Create service file:
`/etc/systemd/system/autoship.service`

```ini
[Unit]
Description=AutoShip Deployment Handler
After=network.target

[Service]
ExecStart=/usr/bin/python3 /opt/autoship/main.py
WorkingDirectory=/opt/autoship
Restart=always

# Logging
StandardOutput=journal
StandardError=journal

# Security / Access
ReadWritePaths=/etc/nginx /etc/letsencrypt \
               /var/lib/autoship/certbot \
               /etc/nginx/sites-available \
               /etc/nginx/sites-enabled \
               /var/www/autoship

# Allow DNS/networking
PrivateNetwork=no
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6

[Install]
WantedBy=multi-user.target
```

Enable and start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable autoship.service
sudo systemctl start autoship.service
```

Check status:

```bash
sudo systemctl status autoship.service --no-pager
```

View logs:

```bash
journalctl -u autoship.service -f
```

---

## 5️⃣ NGINX Integration

* AutoShip generates configs in:
  `/etc/nginx/sites-available/`
  `/etc/nginx/sites-enabled/`

* Example config (test subdomain):

```nginx
server {
    listen 80;
    server_name test.autoship.site;

    location / {
        proxy_pass http://127.0.0.1:8080; # example upstream
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Commands:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

---

## 6️⃣ SSL with Certbot

AutoShip runs Certbot like this:

```bash
sudo certbot --nginx -d <subdomain> \
  --non-interactive --agree-tos -m admin@example.com \
  --no-eff-email \
  --config-dir /var/lib/autoship/certbot/config \
  --work-dir /var/lib/autoship/certbot/work \
  --logs-dir /var/lib/autoship/certbot/logs
```

Manual test:

```bash
sudo certbot --nginx -d test.autoship.site \
  --config-dir /var/lib/autoship/certbot/config \
  --work-dir /var/lib/autoship/certbot/work \
  --logs-dir /var/lib/autoship/certbot/logs
```

---

## 7️⃣ Cloudflare DNS Setup

* Get an API Token with **Zone\:DNS Edit** permissions.
* Store it as environment variable or `.env` for your script:

```bash
export CLOUDFLARE_API_TOKEN="your-api-token"
export CLOUDFLARE_ZONE_ID="your-zone-id"
```

* AutoShip will then create DNS records automatically before issuing SSL.

---

## 8️⃣ Debugging & Maintenance

### Restart service

```bash
sudo systemctl restart autoship.service
```

### Watch logs

```bash
journalctl -u autoship.service -f
```

### Test DNS resolution inside systemd sandbox

```bash
systemd-run --unit=testdns.service --property=PrivateNetwork=no ping -c3 s3.amazonaws.com
```

### Renew SSL manually

```bash
sudo certbot renew \
  --config-dir /var/lib/autoship/certbot/config \
  --work-dir /var/lib/autoship/certbot/work \
  --logs-dir /var/lib/autoship/certbot/logs
```

---

## ✅ Done!

You now have a fully working **AutoShip Deployment Handler** service that:

* Automates NGINX proxy setup
* Manages SSL via Certbot
* Manages DNS via Cloudflare API
* Runs as a persistent systemd service

---

Would you like me to also include a **one-shot bootstrap script** (`setup.sh`) that runs all these folder creations, service setup, and reloads in one go?
