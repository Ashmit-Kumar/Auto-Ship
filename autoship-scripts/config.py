# config.py

# install python-dotenv if not already installed
# pip install python-dotenv

import os
from dotenv import load_dotenv

# Load .env file from root of autoship-host/
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
dotenv_path = os.path.join(BASE_DIR, ".env")
load_dotenv(dotenv_path)

# Access environment variables
GO_APP_HOST = os.getenv("GO_APP_HOST")  # Docker hostname
GO_APP_PORT = os.getenv("GO_APP_PORT")

HOSTINGER_API_KEY = os.getenv("HOSTINGER_API_KEY")
# HOSTINGER_ZONE_ID = os.getenv("HOSTINGER_ZONE_ID")
HOSTINGER_BASE_DOMAIN = os.getenv("HOSTINGER_BASE_DOMAIN")

NGINX_CONF_DIR = os.getenv("NGINX_CONF_DIR", "/etc/nginx/conf.d/")
CERTBOT_EMAIL = os.getenv("CERTBOT_EMAIL")
CERTBOT_WEBROOT = os.getenv("CERTBOT_WEBROOT", "/var/www/certbot")

