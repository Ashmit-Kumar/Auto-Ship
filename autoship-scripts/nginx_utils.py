import os
import subprocess
from config import NGINX_SITES_DIR

NGINX_SITES_DIR = NGINX_SITES_DIR
print(f"NGINX_SITES_DIR: {NGINX_SITES_DIR}")

STATIC_TEMPLATE = """
server {{
    listen 80;
    server_name {subdomain};

    # Redirect all traffic to HTTPS
    return 301 https://$host$request_uri;
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
    }}
}}
"""

DYNAMIC_TEMPLATE = """
server {{
    listen 80;
    server_name {subdomain};

    # Redirect all traffic to HTTPS
    return 301 https://$host$request_uri;
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

def write_nginx_conf_static(subdomain, s3_url):
    conf = STATIC_TEMPLATE.format(subdomain=subdomain, s3_url=s3_url)
    print(f"Writing NGINX conf for static site: {subdomain} with S3 URL: {s3_url}")
    return _write_and_reload(subdomain, conf)

def write_nginx_conf_dynamic(subdomain, port):
    conf = DYNAMIC_TEMPLATE.format(subdomain=subdomain, port=port)
    return _write_and_reload(subdomain, conf)

def _write_and_reload(subdomain, conf_content):
    print("inside _write_and_reload inside nginx_utils.py")
    try:
        path = os.path.join(NGINX_SITES_DIR, f"{subdomain}.conf")
        with open(path, "w") as f:
            f.write(conf_content)
        print(f"[INFO] NGINX conf written to {path}")
        return reload_nginx()
    except Exception as e:
        print(f"[ERROR] Failed to write NGINX conf: {e}")
        return False

def reload_nginx():
    try:
        subprocess.run(["sudo", "nginx", "-t"], check=True)
        subprocess.run(["sudo", "systemctl", "reload", "nginx"], check=True)
        print("[INFO] NGINX reloaded successfully")
        return True
    except subprocess.CalledProcessError as e:
        print(f"[ERROR] NGINX reload failed: {e}")
        return False
