import os
from dotenv import load_dotenv  # Load environment variables from .env file
load_dotenv()
HOSTINGER_API_KEY = os.getenv("HOSTINGER_API_KEY")
HOSTINGER_ZONE_ID = os.getenv("HOSTINGER_ZONE_ID")
NGINX_SITES_DIR=os.getenv("NGINX_SITES_AVAILABLE")
NGINX_SITES_ENABLED=os.getenv("NGINX_SITES_ENABLED")
# LOG_FILE = os.getenv("LOG_FILE", "/var/log/autoship.log")
# DEPLOY_FILE = os.getenv("DEPLOY_FILE", "deploy.json")   