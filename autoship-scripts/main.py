# host_handler.py

import json
import time
import logging
import hashlib
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import os
from nginx_utils import write_nginx_conf_static, write_nginx_conf_dynamic, reload_nginx
from ssl_utils import generate_ssl
from dns_utils import add_dns_record
from response_utils import write_response

# Constants
DEPLOY_FILE = "deploy-requests.json"

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")

last_hash = None
last_event_time = 0
DEBOUNCE_SECONDS = 0.5



def hash_file_content(path):
    try:
        with open(path, 'rb') as f:
            return hashlib.sha256(f.read()).hexdigest()
    except Exception as e:
        logging.warning(f"Hashing failed: {e}")
        return None

def handle_request(req):
    print("Welcome to the host handler! inside handle_request that is inside main.py")
    try:
        req_id = req.get("id")
        subdomain = req.get("subdomain")
        project_type = req.get("project_type")
        if not req_id or not subdomain or not project_type:
            raise ValueError("Missing required fields")

        if project_type == "static":
            s3_url = req.get("s3_url")
            if not s3_url:
                raise ValueError("Missing s3_url for static project")
            write_nginx_conf_static(subdomain, s3_url)

        elif project_type == "dynamic":
            port = req.get("port")
            if not isinstance(port, int):
                raise ValueError("Invalid or missing port for dynamic project")
            write_nginx_conf_dynamic(subdomain, port)

        else:
            raise ValueError("Invalid project_type")

        public_ip = os.getenv("EC2_PUBLIC_IP") # fallback
        if not add_dns_record(subdomain, public_ip):
            raise RuntimeError("DNS record creation failed")

        if not generate_ssl(subdomain):
            raise RuntimeError("SSL certificate generation failed")

        if not reload_nginx():
            raise RuntimeError("NGINX reload failed")

        write_response({
            "request_id": req_id,
            "subdomain": subdomain,
            "status": "success",
            "message": "Deployed",
            **({"s3_url": req.get("s3_url")} if project_type == "static" else {"port": req.get("port")})
        })

    except Exception as e:
        logging.error(f"[{req.get('id')}] Failed: {e}")
        write_response({
            "request_id": req.get("id"),
            "subdomain": req.get("subdomain"),
            "status": "error",
            "message": str(e)
        })


def read_requests_with_retry(retries=3, delay=0.2):
    for attempt in range(retries):
        try:
            with open(DEPLOY_FILE, 'r') as f:
                data = json.load(f)
                return data if isinstance(data, list) else []
        except json.JSONDecodeError:
            logging.warning(f"JSON not ready yet, retrying ({attempt+1}/{retries})...")
            time.sleep(delay)
        except Exception as e:
            logging.error(f"Error reading {DEPLOY_FILE}: {e}")
            return []
    logging.error(f"Failed to read valid JSON after {retries} attempts.")
    return []

TARGET_FILE = Path(DEPLOY_FILE).resolve()

class DeployHandler(FileSystemEventHandler):
    def on_modified(self, event):
        global last_hash, last_event_time
        if event.is_directory:
            return
        if Path(event.src_path).resolve() != TARGET_FILE:
            return
        now = time.time()
        if now - last_event_time < DEBOUNCE_SECONDS:
            logging.debug("Debounced event.")
            return
        last_event_time = now

        new_hash = hash_file_content(TARGET_FILE)
        if not new_hash or new_hash == last_hash:
            logging.debug("No change in file content. Skipping.")
            return

        last_hash = new_hash
        logging.info(f"{DEPLOY_FILE} modified. Reading requests...")
        requests = read_requests_with_retry()
        for req in requests:
            handle_request(req)
        logging.info("Requests processed successfully.")

def main():
    logging.info("Watching for deployment requests...")
    observer = Observer()
    observer.schedule(DeployHandler(), path=str(Path(DEPLOY_FILE).parent), recursive=False)
    observer.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logging.info("Stopping observer.")
        observer.stop()
    observer.join()

if __name__ == "__main__":
    main()



