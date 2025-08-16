#!/bin/bash
set -e  # exit immediately if a command fails

# --- Update system ---
sudo yum update -y

# --- Install Docker ---

sudo yum install -y docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker ec2-user

# --- Install Docker Compose ---
DOCKER_COMPOSE_VERSION=$(curl -s https://api.github.com/repos/docker/compose/releases/latest | grep tag_name | cut -d '"' -f 4)
sudo curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# --- Install Nginx ---

sudo yum install -y nginx
sudo systemctl enable nginx
sudo systemctl start nginx

# --- Run your container ---
# Make sure .env exists in /home/ec2-user/
cd /home/ec2-user/

sudo docker pull ashmit1020/autoship:v1.3.1

# Stop old container if running
if [ "$(sudo docker ps -q -f name=autoship)" ]; then
    sudo docker stop autoship
    sudo docker rm autoship
fi

# Run new container with .env
sudo docker run -d \
  --name autoship \
  --env-file /home/ec2-user/.env \
  -p 5000:5000 \
  ashmit1020/autoship:v1.3.1

echo "✅ Setup complete. Remember to logout/login so ec2-user gets Docker group permissions."
