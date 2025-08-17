Prerequisites

- Ubuntu 22.04 LTS (or similar)
- Docker and Docker Compose
- Go 1.20+ (for building locally) or use Docker build
- Python 3.11+ (for autoship-scripts worker)
- certbot and nginx installed on the host (for TLS)
- MongoDB instance (Atlas or local)

Install Docker & Docker Compose (Ubuntu):

```sh
# Docker
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" |
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Python 3
sudo apt install -y python3 python3-venv python3-pip

# Nginx + certbot
sudo apt install -y nginx certbot python3-certbot-nginx

# Optional: Go
sudo snap install --classic go

# Enable docker for current user
sudo usermod -aG docker $USER
newgrp docker
```
