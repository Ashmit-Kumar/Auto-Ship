import os
from dotenv import load_dotenv

# Load .env file
dotenv_path = os.getenv("DOTENV_PATH", os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env"))
if os.path.exists(dotenv_path):
    load_dotenv(dotenv_path)

# Define environment variables
GO_APP_PROTOCOL = os.getenv("GO_APP_PROTOCOL", "http")
GO_APP_HOST = os.getenv("GO_APP_HOST")
GO_APP_PORT = os.getenv("GO_APP_PORT")
GO_APP_REQUEST_TIMEOUT = float(os.getenv("GO_APP_REQUEST_TIMEOUT", "10"))
HOSTINGER_API_TOKEN = os.getenv("HOSTINGER_API_TOKEN")
HOSTINGER_API_BASE_URL = os.getenv("HOSTINGER_API_BASE_URL", "https://api.hostinger.com/v1")
DNS_TTL = int(os.getenv("DNS_TTL", "300"))
EC2_PUBLIC_IP = os.getenv("EC2_PUBLIC_IP")
NGINX_SITES_AVAILABLE = os.getenv("NGINX_SITES_AVAILABLE", "/etc/nginx/sites-available")
NGINX_SITES_ENABLED = os.getenv("NGINX_SITES_ENABLED", "/etc/nginx/sites-enabled")
CERTBOT_EMAIL = os.getenv("CERTBOT_EMAIL", "admin@example.com")
LOG_FILE = os.getenv("LOG_FILE", "/var/log/autoship.log")

def validate_env_vars():
    """Validate required environment variables."""
    required = [
        "GO_APP_HOST",
        "GO_APP_PORT",
        "HOSTINGER_API_TOKEN",
        "EC2_PUBLIC_IP",
        "CERTBOT_EMAIL"
    ]
    missing = [var for var in required if not os.getenv(var)]
    if missing:
        raise ValueError(f"Missing environment variables: {missing}")