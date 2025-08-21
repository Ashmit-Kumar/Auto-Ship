# Run autoship worker as a systemd service

This document explains how to install and run the host worker scripts under systemd so the worker starts on boot and restarts on failure.

Prerequisites
- The worker files copied to `/opt/autoship` (scripts like `main.py`, `dns_utils.py`, etc.).
- An environment file at `/opt/autoship/.env` containing required secrets (Cloudflare token, zone id, EC2 IP, nginx paths).
- Python 3 installed on the host and any required packages available system-wide or in `/opt/autoship/venv`.

1) Create target directory and copy files

Run as root or with sudo:

sudo mkdir -p /opt/autoship
sudo cp -r /home/ashmit/code/Auto-Ship/autoship-scripts/* /opt/autoship/

2) Create `.env`

Create `/opt/autoship/.env` and add values (replace placeholders):

```
CLOUDFLARE_API_TOKEN=your_token_here
CLOUDFLARE_ZONE_ID=your_zone_id_here
EC2_PUBLIC_IP=your_server_ip
NGINX_SITES_AVAILABLE=/etc/nginx/sites-available
NGINX_SITES_ENABLED=/etc/nginx/sites-enabled
```

Secure the file:

sudo chown root:root /opt/autoship/.env
sudo chmod 600 /opt/autoship/.env

3) (Optional) Create and activate a virtualenv

If you prefer a virtualenv, create one at `/opt/autoship/venv` and install requirements:

sudo python3 -m venv /opt/autoship/venv
sudo /opt/autoship/venv/bin/pip install -r /opt/autoship/requirements.txt

Adjust the `ExecStart` below to use the venv Python (`/opt/autoship/venv/bin/python`).

4) Create the systemd service file

Create `/etc/systemd/system/autoship.service` with the following content (paste as root):

```
[Unit]
Description=AutoShip Deployment Handler
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/python3 /opt/autoship/main.py
WorkingDirectory=/opt/autoship
EnvironmentFile=/opt/autoship/.env
Restart=on-failure
User=root

# Allow write access to /etc/nginx and certbot data directories which the worker
# and certbot need to modify. ProtectSystem=full makes /etc read-only and will
# cause errors like "PluginError('Unable to lock /etc/nginx')". Use ReadWritePaths
# to whitelist directories that must remain writable.
ReadWritePaths=/etc/nginx /var/lib/autoship/certbot /etc/letsencrypt

# Security hardening (optional)
# ProtectSystem=full  # DO NOT enable while certbot/nginx need write access
ProtectHome=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

Notes:
- If you created a venv, change `ExecStart` to `/opt/autoship/venv/bin/python /opt/autoship/host_handler.py`.
- Running as `root` is convenient because the worker needs to write nginx files, run certbot and reload services; if you want to use a non-root user, grant that user appropriate sudo privileges for the specific commands (nginx reload, certbot) and adjust `User=` accordingly.



5) Prepare certbot directories and permissions (fixes "Unable to lock /etc/nginx")

When running certbot from a systemd service that uses ProtectSystem or other sandboxing, ensure certbot can write its config and logs. Run as root:

```bash
sudo mkdir -p /var/lib/autoship/certbot/{config,work,logs}
sudo chown -R root:root /var/lib/autoship/certbot
sudo chmod -R 700 /var/lib/autoship/certbot
sudo mkdir -p /etc/letsencrypt
sudo chown root:root /etc/letsencrypt
```

If you see errors like "The nginx plugin is not working; PluginError('Unable to lock /etc/nginx')", update the service to include `ReadWritePaths=/etc/nginx` (or remove `ProtectSystem=full`) and then reload systemd and restart the service:

```bash
sudo systemctl daemon-reload
sudo systemctl restart autoship.service
sudo journalctl -u autoship.service -f
```

6) Enable and start the service

sudo systemctl daemon-reload
sudo systemctl enable autoship.service
sudo systemctl start autoship.service

7) Verify status and view logs

Check status:

sudo systemctl status autoship.service

Follow live logs:

journalctl -u autoship.service -f

If you prefer logging to a file, add to the `[Service]` section:

StandardOutput=append:/var/log/autoship.log
StandardError=append:/var/log/autoship.log

Then create and secure the log file:

sudo touch /var/log/autoship.log
sudo chown root:root /var/log/autoship.log
sudo chmod 640 /var/log/autoship.log

8) Common troubleshooting
- "No such file or directory" for `/opt/autoship/.env`: verify file exists and permissions allow the service to read it.
- Certbot failures: ensure the domain resolves to the EC2 public IP and ports 80/443 are allowed by the firewall/EC2 security group.
- Permissions for nginx files: either run as root or allow the service user to write under `/etc/nginx/sites-available` and run `nginx -s reload` via sudo.

9) Auto renew SSL
- Certbot typically installs a systemd timer for renewals. Verify with `systemctl list-timers | grep certbot`.
- You can add a daily systemd timer or cron job that runs `certbot renew --quiet && systemctl reload nginx`.

10) Service security recommendations (optional)
- Use `User=autoship` (non-root) and give that user limited sudo rights for the exact commands the worker needs.
- Limit file system access with `ReadOnlyPaths=` / `InaccessiblePaths=` if possible.
- Keep secrets in `/opt/autoship/.env` with restrictive permissions (600).

If you want, I can:
- Generate the exact `/etc/systemd/system/autoship.service` file and a matching sudoers snippet for a non-root `autoship` user.
- Add an example `requirements.txt` and a sample `host_handler.py` wrapper that logs to `/var/log/autoship.log`.

---

Document created by automation — follow the steps above on your EC2 host to run the worker as a systemd service.
