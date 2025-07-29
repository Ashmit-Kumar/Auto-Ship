# response_utils.py

import requests
import logging
import os

logging.basicConfig(level=logging.INFO)

# Configure this to match your Go app's container hostname or Docker network alias
GO_APP_HOST = os.getenv("GO_APP_HOST")
GO_APP_PORT = os.getenv("GO_APP_PORT")  # or whatever your app uses

def send_status_to_go_app(request_id: str, subdomain: str, success: bool, message: str = "", port: int = None, s3_url: str = None):
    """
    Sends the deployment result back to the Go app.
    """
    url = f"http://{GO_APP_HOST}:{GO_APP_PORT}/api/deployment-status"
    payload = {
        "request_id": request_id,
        "subdomain": subdomain,
        "status": "success" if success else "failed",
        "message": message,
    }

    if port:
        payload["port"] = port
    if s3_url:
        payload["s3_url"] = s3_url

    try:
        logging.info(f"Sending deployment status for {subdomain} to Go app...")
        response = requests.post(url, json=payload, timeout=10)

        if response.status_code == 200:
            logging.info("Status successfully sent to Go app.")
            return True
        else:
            logging.error(f"Failed to send status to Go app: {response.status_code} - {response.text}")
            return False

    except requests.RequestException as e:
        logging.error(f"Exception while sending status to Go app: {e}")
        return False
