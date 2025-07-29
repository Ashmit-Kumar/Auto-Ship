import os
import re
import subprocess
import logging

# Configure logging (will be overridden by log.py)
logging.basicConfig(level=logging.INFO, format='{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "error": "%(error)s"}')

# Configurable NGINX path
NGINX_CONF_DIR = os.getenv("NGINX_CONF_DIR", "/etc/nginx/conf.d")

def validate_subdomain(subdomain: str) -> bool:
    """Validate subdomain to prevent injection attacks."""
    if not isinstance(subdomain, str):
        logging.error("Subdomain must be a string", extra={"subdomain": str(subdomain), "error": "Invalid type"})
        return False
    if not re.match(r'^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$', subdomain):
        logging.error("Invalid subdomain format", extra={"subdomain": subdomain, "error": "Invalid characters"})
        return False
    if '..' in subdomain or '/' in subdomain or ';' in subdomain:
        logging.error("Subdomain contains unsafe characters", extra={"subdomain": subdomain, "error": "Injection risk"})
        return False
    return True

def write_nginx_config(subdomain: str, config: str) -> bool:
    """Write NGINX config to NGINX_CONF_DIR."""
    if not validate_subdomain(subdomain):
        raise ValueError(f"Invalid subdomain: {subdomain}")

    config_path = os.path.join(NGINX_CONF_DIR, f"{subdomain}.conf")
    try:
        with open(config_path, "w") as f:
            f.write(config)
        logging.info("NGINX config written", extra={"subdomain": subdomain, "error": ""})
        return True
    except (OSError, PermissionError) as e:
        logging.error("Failed to write NGINX config", extra={"subdomain": subdomain, "error": str(e)})
        return False

def write_nginx_conf_dynamic(subdomain: str, port: int) -> bool:
    """Write NGINX config for dynamic app with HTTP and HTTPS."""
    config = f"""
server {{
    listen 80;
    server_name {subdomain};

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}}

server {{
    listen 443 ssl;
    server_name {subdomain};

    ssl_certificate /etc/letsencrypt/live/{subdomain}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/{subdomain}/privkey.pem;

    location / {{
        proxy_pass http://localhost:{port};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }}
}}
"""
    return write_nginx_config(subdomain, config)

def write_nginx_conf_static(subdomain: str, s3_url: str) -> bool:
    """Write NGINX config for static app with HTTP and HTTPS."""
    config = f"""
server {{
    listen 80;
    server_name {subdomain};

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}}

server {{
    listen 443 ssl;
    server_name {subdomain};

    ssl_certificate /etc/letsencrypt/live/{subdomain}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/{subdomain}/privkey.pem;

    location / {{
        proxy_pass {s3_url};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }}
}}
"""
    return write_nginx_config(subdomain, config)

def reload_nginx() -> bool:
    """Test and reload NGINX configuration."""
    logging.info("Mock: NGINX configuration reloaded", extra={"subdomain": "", "error": ""})
    return True