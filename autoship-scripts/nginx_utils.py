import os
import re
import subprocess
import logging

# Configure logging
logging.basicConfig(level=logging.INFO, format='{"time": "%(asctime)s", "level": "%(levelname)s", "message"), "subdomain": "%(subdomain)s", "error": "%(error)s"}')

# Configurable NGINX paths
NGINX_SITES_AVAILABLE = os.getenv("NGINX_SITES_AVAILABLE", "/etc/nginx/sites-available")
NGINX_SITES_ENABLED = os.getenv("NGINX_SITES_ENABLED", "/etc/nginx/sites-enabled")

def validate_subdomain(subdomain: str) -> bool:
    """Validate subdomain to prevent injection attacks."""
    if not isinstance(subdomain, str):
        logging.error("Subdomain must be a string", extra={"subdomain": str(subdomain), "error": "Invalid type"})
        return False
    if not re.match(r'^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$', subdomain):
        logging.error("Invalid subdomain format", extra={"subdomain": subdomain, "error": "Invalid characters"})
        return False
    if '..' in subdomain or '/' in subdomain:
        logging.error("Subdomain contains unsafe characters", extra={"subdomain": subdomain, "error": "Path traversal risk"})
        return False
    return True

def write_nginx_config(subdomain: str, config: str) -> bool:
    """Write NGINX config and create symlink."""
    if not validate_subdomain(subdomain):
        raise ValueError(f"Invalid subdomain: {subdomain}")

    config_path = os.path.join(NGINX_SITES_AVAILABLE, subdomain)
    enabled_path = os.path.join(NGINX_SITES_ENABLED, subdomain)

    try:
        # Write config file
        with open(config_path, "w") as f:
            f.write(config)
        logging.info("NGINX config written", extra={"subdomain": subdomain, "error": ""})

        # Create or update symlink
        if os.path.exists(enabled_path):
            if os.path.islink(enabled_path) and os.readlink(enabled_path) == config_path:
                logging.info("Symlink already correct", extra={"subdomain": subdomain, "error": ""})
            else:
                os.remove(enabled_path)
                os.symlink(config_path, enabled_path)
                logging.info("Symlink updated", extra={"subdomain": subdomain, "error": ""})
        else:
            os.symlink(config_path, enabled_path)
            logging.info("Symlink created", extra={"subdomain": subdomain, "error": ""})
        return True
    except (OSError, PermissionError) as e:
        logging.error("Failed to write NGINX config or create symlink", extra={"subdomain": subdomain, "error": str(e)})
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
    try:
        result = subprocess.run(["nginx", "-t"], capture_output=True, text=True)
        if result.returncode != 0:
            logging.error("NGINX config test failed", extra={"subdomain": "", "error": result.stderr})
            raise RuntimeError(f"NGINX config test failed: {result.stderr}")
        result = subprocess.run(["nginx", "-s", "reload"], capture_output=True, text=True)
        if result.returncode == 0:
            logging.info("NGINX configuration reloaded", extra={"subdomain": "", "error": ""})
            return True
        else:
            logging.error("NGINX reload failed", extra={"subdomain": "", "error": result.stderr})
            return False
    except (OSError, subprocess.SubprocessError) as e:
        logging.error("Error reloading NGINX", extra={"subdomain": "", "error": str(e)})
        return False


# # nginx_utils.py
# import os

# def write_nginx_conf_dynamic(subdomain, port):
#     config = f"""
#     server {{
#         listen 80;
#         server_name {subdomain};

#         location / {{
#             proxy_pass http://localhost:{port};
#             proxy_http_version 1.1;
#             proxy_set_header Upgrade $http_upgrade;
#             proxy_set_header Connection 'upgrade';
#             proxy_set_header Host $host;
#             proxy_cache_bypass $http_upgrade;
#         }}
#     }}
#     """
#     with open(f"/etc/nginx/sites-available/{subdomain}", "w") as f:
#         f.write(config)
#     os.system(f"ln -sf /etc/nginx/sites-available/{subdomain} /etc/nginx/sites-enabled/{subdomain}")

# def write_nginx_conf_static(subdomain, s3_url):
#     config = f"""
#     server {{
#         listen 80;
#         server_name {subdomain};

#         location / {{
#             proxy_pass {s3_url};
#         }}
#     }}
#     """
#     with open(f"/etc/nginx/sites-available/{subdomain}", "w") as f:
#         f.write(config)
#     os.system(f"ln -sf /etc/nginx/sites-available/{subdomain} /etc/nginx/sites-enabled/{subdomain}")


# def reload_nginx():
#     result = subprocess.run(["nginx", "-t"], capture_output=True, text=True)
#     if result.returncode != 0:
#         logging.error(f"NGINX config test failed: {result.stderr}")
#         return False
#     subprocess.run(["nginx", "-s", "reload"])
#     logging.info("NGINX configuration reloaded.")
#     return True