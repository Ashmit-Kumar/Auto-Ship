import requests
import logging
import os
import re
from tenacity import retry, stop_after_attempt, wait_exponential

# Configure structured logging
logging.basicConfig(level=logging.INFO, format='{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "request_id": "%(request_id)s", "error": "%(error)s"}')

# Configurable Go app settings
GO_APP_PROTOCOL = os.getenv("GO_APP_PROTOCOL", "http")
GO_APP_HOST = os.getenv("GO_APP_HOST")
GO_APP_PORT = os.getenv("GO_APP_PORT")
REQUEST_TIMEOUT = float(os.getenv("GO_APP_REQUEST_TIMEOUT", "10"))

# Validate environment variables
if not GO_APP_HOST or not GO_APP_PORT:
    raise ValueError("GO_APP_HOST and GO_APP_PORT environment variables are required")

def validate_string_param(param: str, name: str, allow_empty: bool = False) -> bool:
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
        logging.error(f"{name} contains unsafe characters", extra={"subdomain": param if name == "subdomain" else "", "request_id": param if name == "request_id" else "", "error": "Injection risk"})
        return False
    return True

@retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=2, max=10))
def send_status_to_go_app(request_id: str, subdomain: str, success: bool, message: str = "", port: int = None, s3_url: str = None) -> bool:
    """
    Sends the deployment result back to the Go app.
    """
    # Validate inputs
    if not validate_string_param(request_id, "request_id"):
        raise ValueError(f"Invalid request_id: {request_id}")
    if not validate_string_param(subdomain, "subdomain"):
        raise ValueError(f"Invalid subdomain: {subdomain}")
    if not validate_string_param(message, "message", allow_empty=True):
        raise ValueError(f"Invalid message: {message}")
    if port is not None and not isinstance(port, int):
        logging.error("Port must be an integer", extra={"subdomain": subdomain, "request_id": request_id, "error": f"Invalid type: {type(port)}"})
        raise ValueError(f"Invalid port: {port}")
    if s3_url is not None and not isinstance(s3_url, str):
        logging.error("S3 URL must be a string", extra={"subdomain": subdomain, "request_id": request_id, "error": f"Invalid type: {type(s3_url)}"})
        raise ValueError(f"Invalid s3_url: {s3_url}")

    url = f"{GO_APP_PROTOCOL}://{GO_APP_HOST}:{GO_APP_PORT}/api/deployment-status"
    payload = {
        "request_id": request_id,
        "subdomain": subdomain,
        "status": "success" if success else "failed",
        "message": message,
    }
    if port is not None:
        payload["port"] = port
    if s3_url:
        payload["s3_url"] = s3_url

    try:
        logging.info("Sending deployment status to Go app", extra={"subdomain": subdomain, "request_id": request_id, "error": ""})
        response = requests.post(url, json=payload, timeout=REQUEST_TIMEOUT)

        if response.status_code == 200:
            logging.info("Status successfully sent to Go app", extra={"subdomain": subdomain, "request_id": request_id, "error": ""})
            return True
        else:
            try:
                error_msg = response.json()
            except ValueError:
                error_msg = response.text
            logging.error("Failed to send status to Go app", extra={"subdomain": subdomain, "request_id": request_id, "error": f"{response.status_code} - {error_msg}"})
            return False
    except requests.RequestException as e:
        logging.error("Network error sending status to Go app", extra={"subdomain": subdomain, "request_id": request_id, "error": str(e)})
        return False
    except Exception as e:
        logging.error("Unexpected error sending status to Go app", extra={"subdomain": subdomain, "request_id": request_id, "error": str(e)})
        return False