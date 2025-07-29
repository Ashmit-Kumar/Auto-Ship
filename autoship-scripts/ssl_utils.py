# ssl_utils.py

import subprocess
import logging

logging.basicConfig(level=logging.INFO)

def generate_ssl(subdomain: str) -> bool:
    """
    Generates SSL certificate for a subdomain using Certbot with NGINX plugin.
    """
    try:
        logging.info(f"Generating SSL certificate for: {subdomain}")
        result = subprocess.run(
            ["sudo", "certbot", "--nginx", "-d", subdomain, "--non-interactive", "--agree-tos", "-m", "admin@" + subdomain],
            capture_output=True,
            text=True
        )
        if result.returncode == 0:
            logging.info(f"SSL generation successful for {subdomain}")
            return True
        else:
            logging.error(f"Certbot failed: {result.stderr}")
            return False
    except Exception as e:
        logging.error(f"Error generating SSL for {subdomain}: {e}")
        return False
