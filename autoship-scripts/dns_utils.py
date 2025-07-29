import requests
import logging
import os
import re
from tenacity import retry, stop_after_attempt, wait_exponential

# Configure structured logging
logging.basicConfig(level=logging.INFO, format='{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "error": "%(error)s"}')

# Configurable API settings
HOSTINGER_API_TOKEN = os.getenv("HOSTINGER_API_TOKEN")
HOSTINGER_API_BASE_URL = os.getenv("HOSTINGER_API_BASE_URL", "https://api.hostinger.com/v1")
DNS_TTL = int(os.getenv("DNS_TTL", "300"))

# Validate API token
if not HOSTINGER_API_TOKEN:
    raise ValueError("HOSTINGER_API_TOKEN environment variable is required")

HEADERS = {
    "Authorization": f"Bearer {HOSTINGER_API_TOKEN}",
    "Content-Type": "application/json"
}

def validate_subdomain(subdomain: str) -> bool:
    """Validate subdomain to prevent invalid or unsafe inputs."""
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

@retry(stop=stop_after_attempt(3), wait=wait_exponential(multiplier=1, min=2, max=10))
def add_dns_record(subdomain: str, public_ip: str) -> bool:
    """
    Adds an A record for subdomain -> public_ip using Hostinger DNS API.
    """
    if not validate_subdomain(subdomain):
        raise ValueError(f"Invalid subdomain: {subdomain}")

    try:
        parts = subdomain.split(".")
        if len(parts) < 3:
            logging.error("Subdomain must be in the format sub.domain.tld", extra={"subdomain": subdomain, "error": "Invalid format"})
            return False

        domain = ".".join(parts[-2:])
        record_name = ".".join(parts[:-2]) or "@"  # Use "@" for root subdomain

        payload = {
            "type": "A",
            "name": record_name,
            "value": public_ip,
            "ttl": DNS_TTL
        }

        logging.info(f"Adding A record for {subdomain} → {public_ip}", extra={"subdomain": subdomain, "error": ""})
        response = requests.post(
            f"{HOSTINGER_API_BASE_URL}/domains/{domain}/records",
            headers=HEADERS,
            json=payload
        )

        if response.status_code == 201:
            logging.info("DNS record created", extra={"subdomain": subdomain, "error": ""})
            return True
        else:
            try:
                error_msg = response.json()
            except ValueError:
                error_msg = response.text
            logging.error("Failed to create DNS record", extra={"subdomain": subdomain, "error": str(error_msg)})
            return False
    except requests.RequestException as e:
        logging.error("Network error in add_dns_record", extra={"subdomain": subdomain, "error": str(e)})
        return False
    except Exception as e:
        logging.error("Unexpected error in add_dns_record", extra={"subdomain": subdomain, "error": str(e)})
        return False