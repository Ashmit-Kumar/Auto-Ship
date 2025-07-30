import json
import os
import time
import logging
from pathlib import Path
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import fcntl
import re


from config import validate_env_vars
from nginx_utils import write_nginx_conf_dynamic, write_nginx_conf_static, reload_nginx
from ssl_utils import generate_ssl, check_dns_propagation
from dns_utils import add_dns_record
from response_utils import send_status_to_go_app as send_status
import log  # Import centralized logging configuration

DEPLOY_FILE = read_json_file("deploy-requests.json")
# DEPLOY_FILE_BACKUP = read_json_file("deploy-requests.json.bak")
PROCESSED_FILE = read_json_file("processed.json")
PROCESSED_BACKUP_FILE = read_json_file("processed.json.bak")

def read_json_file(file_path):
    if not os.path.exists(file_path):
        return []

    with open(file_path, 'r') as f:
        content = f.read().strip()
        if not content:
            return []  # Return empty list if file is empty
        return json.loads(content)


def validate_string_param(param: str, name: str, allow_empty: bool = False) -> bool:
    print(f"Welcome to validate_string_param! This function will validate the {name} parameter.")
    """Validate string parameters to prevent invalid or unsafe inputs."""
    if not isinstance(param, str):
        logging.error(f"{name} must be a string", extra={"subdomain": "", "request_id": "", "error": f"Invalid type: {type(param)}"})
        return False
    if not allow_empty and not param:
        logging.error(f"{name} cannot be empty", extra={"subdomain": "", "request_id": "", "error": "Empty string"})
        return False
    if not re.match(r'^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$', param) and param:
        logging.error(f"Invalid {name} format", extra={"subdomain": param if name == "subdomain" else "", "request_id": param if name == "request_id" else "", "error": "Invalid characters"})
        return False
    if '..' in param or '/' in param or ';' in param:
        print(f"{name} contains unsafe characters: {param}")
        logging.error(f"{name} contains unsafe characters", extra={"subdomain": param if name == "subdomain" else "", "request_id": param if name == "request_id" else "", "error": "Injection risk"})
        return False
    return True

def safe_read_json_file(path):
    print("Welcome to safe_read_json_file! This function will read a JSON file with file locking.")
    """Read JSON file with file locking."""
    print(f"Reading JSON file: {path}")
    if not os.path.exists(path):
        logging.warning(f"File {path} does not exist", extra={"subdomain": "", "request_id": "", "error": "File not found"})
        return []
    try:
        with open(path, "r") as f:
            fcntl.flock(f.fileno(), fcntl.LOCK_SH)
            try:
                raw = f.read().strip()
                if not raw:
                    raise ValueError("File is empty")
                return json.loads(raw)
            finally:
                fcntl.flock(f.fileno(), fcntl.LOCK_UN)
    except (OSError, ValueError, json.JSONDecodeError) as e:
        logging.error(f"Error reading JSON file {path}", extra={"subdomain": "", "request_id": "", "error": str(e)})
        raise ValueError(f"Error reading JSON file {path}: {e}")

def load_json(path):
    print("Welcome to load_json! This function will load a JSON file or return an empty list if the file doesn't exist.")
    """Load JSON file or return empty list if file doesn't exist."""
    return safe_read_json_file(path) if os.path.exists(path) else []

def save_json(path, data):
    print("Welcome to save_json! This function will save JSON data with file locking and backup.")
    """Save JSON data with file locking and backup."""
    try:
        # Create backup of existing file
        if os.path.exists(path):
            os.replace(path, f"{path}.bak")
        with open(path, "w") as f:
            fcntl.flock(f.fileno(), fcntl.LOCK_EX)
            try:
                json.dump(data, f, indent=2)
            finally:
                fcntl.flock(f.fileno(), fcntl.LOCK_UN)
        logging.info(f"Saved JSON to {path}", extra={"subdomain": "", "request_id": "", "error": ""})
    except (OSError, json.JSONDecodeError) as e:
        logging.error(f"Error saving JSON to {path}", extra={"subdomain": "", "request_id": "", "error": str(e)})
        # Restore backup if save fails
        if os.path.exists(f"{path}.bak"):
            os.replace(f"{path}.bak", path)
        raise ValueError(f"Error saving JSON to {path}: {e}")

def mark_as_processed(processed_ids, req_id):
    print("Welcome to mark_as_processed! This function will mark a request as processed.")
    """Mark request as processed and limit to 1000 entries."""
    processed_ids.append(req_id)
    if len(processed_ids) > 1000:
        processed_ids = processed_ids[-1000:]
    save_json(PROCESSED_FILE, processed_ids)

def handle_request(req, processed_ids):
    print("Welcome to handle_request! This function will process a single deployment request.")
    """Process a single deployment request."""
    req_id = req.get("id")
    subdomain = req.get("subdomain")
    project_type = req.get("project_type")

    # Validate inputs
    if not validate_string_param(req_id, "request_id"):
        logging.warning(f"Invalid request_id: {req}", extra={"subdomain": subdomain or "", "request_id": req_id or "", "error": "Invalid request_id"})
        return
    if not validate_string_param(subdomain, "subdomain"):
        logging.warning(f"Invalid subdomain: {req}", extra={"subdomain": subdomain or "", "request_id": req_id, "error": "Invalid subdomain"})
        return
    if not validate_string_param(project_type, "project_type") or project_type not in ["static", "dynamic"]:
        logging.warning(f"Invalid project_type: {req}", extra={"subdomain": subdomain, "request_id": req_id, "error": f"Invalid project_type: {project_type}"})
        return

    if req_id in processed_ids:
        logging.info(f"Request {req_id} already processed", extra={"subdomain": subdomain, "request_id": req_id, "error": ""})
        return

    logging.info(f"Processing this request in handle_request: {subdomain} ({project_type})", extra={"subdomain": subdomain, "request_id": req_id, "error": ""})
    try:
        if project_type == "dynamic":
            port = req.get("port")
            if not isinstance(port, int):
                raise ValueError(f"Invalid port: {port}")
            if not write_nginx_conf_dynamic(subdomain, port):
                raise RuntimeError("Failed to write NGINX config for dynamic app")
        elif project_type == "static":
            if not subdomain.endswith(".site"):
                raise ValueError("Subdomain must end with .site for static apps")
            s3_url = req.get("s3_url")
            if not isinstance(s3_url, str) or not s3_url:
                raise ValueError(f"Invalid s3_url: {s3_url}")
            if not write_nginx_conf_static(subdomain, s3_url):
                raise RuntimeError("Failed to write NGINX config for static app")

        public_ip = os.getenv("EC2_PUBLIC_IP")
        if not public_ip:
            raise ValueError("EC2_PUBLIC_IP environment variable is missing")
        if not add_dns_record(subdomain, public_ip):
            raise RuntimeError(f"Failed to add DNS record for {subdomain}")
        if not check_dns_propagation(subdomain, public_ip):
            raise RuntimeError(f"DNS propagation failed for {subdomain}")
        if not generate_ssl(subdomain):
            raise RuntimeError(f"Failed to generate SSL for {subdomain}")
        if not reload_nginx():
            raise RuntimeError("Failed to reload NGINX")
        send_status(req_id, subdomain=subdomain, success=True, message="Deployed")
        mark_as_processed(processed_ids, req_id)
        logging.info(f"Success: {subdomain}", extra={"subdomain": subdomain, "request_id": req_id, "error": ""})
    except Exception as e:
        logging.error(f"Error processing {subdomain}", extra={"subdomain": subdomain, "request_id": req_id, "error": str(e)})
        send_status(req_id, subdomain=subdomain, success=False, message=str(e))

class DeployFileHandler(FileSystemEventHandler):
    """Handle file modifications for deploy-requests.json."""
    def on_modified(self, event):
        print("Welcome to DeployFileHandler! This class handles file modifications for deploy-requests.json.")
        if event.src_path.endswith(DEPLOY_FILE):
            logging.info("deploy-requests.json modified", extra={"subdomain": "", "request_id": "", "error": ""})
            process_requests()

def process_requests():
    print("Welcome to process requests! function This function will process all requests in deploy-requests.json.")
    """Process all requests in deploy-requests.json."""
    try:
        requests = safe_read_json_file(DEPLOY_FILE)
        if isinstance(requests, dict):
            requests = [requests]
        processed_ids = load_json(PROCESSED_FILE)
        for req in requests:
            handle_request(req, processed_ids)
    except ValueError as e:
        logging.error("Failed to process requests", extra={"subdomain": "", "request_id": "", "error": str(e)})

def main():
    """Main loop to watch for deployment requests."""
    validate_env_vars()
    logging.info("Watching for deployment requests...", extra={"subdomain": "", "request_id": "", "error": ""})
    observer = Observer()
    observer.schedule(DeployFileHandler(), path=DEPLOY_FILE, recursive=False)
    observer.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        logging.info("Shutting down observer...", extra={"subdomain": "", "request_id": "", "error": ""})
        observer.stop()
    observer.join()

if __name__ == "__main__":
    main()