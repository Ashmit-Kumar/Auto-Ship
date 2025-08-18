EC2 deployment (recommended setup)

1. Create EC2 instance (Ubuntu 22.04), allow ports 22, 80, 443, and the host port range you plan to expose (2000-65535) in security group.

2. Install dependencies (see INSTALL.md)

3. Clone repo and prepare .env

4. Create shared folder for server-worker communication

```sh
sudo mkdir -p /var/lib/autoship/deploy
sudo chown ubuntu:ubuntu /var/lib/autoship/deploy
sudo chmod 775 /var/lib/autoship/deploy
```

5. Start MongoDB (Atlas recommended). If local, run as docker container.

6. Run Go server inside a container (mount the shared folder):

```shn# example using docker
docker build -t autoship-server:latest -f autoship-server/dockerfile ./autoship-server
docker run -d --name autoship-server -p 5000:5000 \
  -v /var/lib/autoship/deploy:/var/lib/autoship/deploy \
  --env-file ./autoship-server/.env autoship-server:latest
```

7. Run Python worker on the host (systemd)

Create systemd file /etc/systemd/system/autoship-worker.service:

````ini
[Unit]
Description=Autoship Host Worker
After=network.target

[Service]
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu/Auto-Ship/autoship-scripts
ExecStart=/home/ubuntu/Auto-Ship/autoship-scripts/venv/bin/python /home/ubuntu/Auto-Ship/autoship-scripts/main.py
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
````

Then enable:

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now autoship-worker
sudo journalctl -u autoship-worker -f
```

8. Verify deployment flow by POSTing to server /api/projects and check /var/lib/autoship/deploy/deploy-responses.json

9. To delete deployments, call DELETE /projects/:containerName which removes container and deletes project document.

