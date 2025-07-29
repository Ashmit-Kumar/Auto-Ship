# nginx_utils.py
import os

def write_nginx_conf_dynamic(subdomain, port):
    config = f"""
    server {{
        listen 80;
        server_name {subdomain};

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
    with open(f"/etc/nginx/sites-available/{subdomain}", "w") as f:
        f.write(config)
    os.system(f"ln -sf /etc/nginx/sites-available/{subdomain} /etc/nginx/sites-enabled/{subdomain}")

def write_nginx_conf_static(subdomain, s3_url):
    config = f"""
    server {{
        listen 80;
        server_name {subdomain};

        location / {{
            proxy_pass {s3_url};
        }}
    }}
    """
    with open(f"/etc/nginx/sites-available/{subdomain}", "w") as f:
        f.write(config)
    os.system(f"ln -sf /etc/nginx/sites-available/{subdomain} /etc/nginx/sites-enabled/{subdomain}")


def reload_nginx():
    os.system("nginx -s reload")
    print("Nginx configuration reloaded.")

