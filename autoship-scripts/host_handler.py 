# host_handler.py

import json
import time
import os

from config import *
from nginx_utils import write_nginx_conf_dynamic, write_nginx_conf_static, reload_nginx
from ssl_utils import generate_ssl
from dns_utils import add_dns_record
from response_utils import send_status
from pathlib import Path

DEPLOY_FILE = "deploy-requests.json"
PROCESSED_FILE = "processed.json"

def load_json(path):
    if not os.path.exists(path):
        return []
    with open(path, "r") as f:
        return json.load(f)

def save_json(path, data):
    with open(path, "w") as f:
        json.dump(data, f, indent=2)

def mark_as_processed(processed_ids, req_id):
    processed_ids.append(req_id)
    save_json(PROCESSED_FILE, processed_ids)

def handle_request(req, processed_ids):
    req_id = req.get("id")
    subdomain = req.get("subdomain")
    project_type = req.get("project_type")  # "static" or "dynamic"

    if not req_id or not subdomain or not project_type:
        print(f"[WARN] Invalid request: {req}")
        return

    if req_id in processed_ids:
        return

    print(f"[INFO] Processing: {subdomain} ({project_type})")

    try:
        if project_type == "dynamic":
            port = req.get("port")
            if not port:
                raise ValueError("Port not provided for dynamic app")

            write_nginx_conf_dynamic(subdomain, port)
        elif project_type == "static":
            s3_url = req.get("s3_url")
            if not s3_url:
                raise ValueError("s3_url not provided for static app")

            write_nginx_conf_static(subdomain, s3_url)
        else:
            raise ValueError("Unknown project_type")

        # Add DNS record
        add_dns_record(subdomain)

        # Generate SSL certificate
        generate_ssl(subdomain)

        # Reload NGINX
        reload_nginx()

        # Notify Go app
        send_status(req_id, status="success", message="Deployed", subdomain=subdomain)

        # Mark processed
        mark_as_processed(processed_ids, req_id)

        print(f"[✓] Success: {subdomain}")

    except Exception as e:
        print(f"[✗] Error processing {subdomain}: {e}")
        send_status(req_id, status="error", message=str(e), subdomain=subdomain)

def main():
    print("[HOST] Watching for deployment requests...")

    last_data = ""
    while True:
        try:
            if not Path(DEPLOY_FILE).exists():
                time.sleep(1)
                continue

            with open(DEPLOY_FILE, "r") as f:
                data = f.read()

            if data == last_data:
                time.sleep(1)
                continue

            last_data = data
            requests = json.loads(data)
            processed_ids = load_json(PROCESSED_FILE)

            for req in requests:
                handle_request(req, processed_ids)

        except json.JSONDecodeError:
            print("[WARN] deploy-requests.json not ready or corrupted.")
        except Exception as e:
            print(f"[ERROR] {e}")
        
        time.sleep(1)

if __name__ == "__main__":
    main()
