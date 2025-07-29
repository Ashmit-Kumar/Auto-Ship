# dns_utils.py

import requests
import logging
import os

logging.basicConfig(level=logging.INFO)

# Set your Hostinger API token in environment
HOSTINGER_API_TOKEN = os.getenv("HOSTINGER_API_TOKEN")
HOSTINGER_API_URL = "https://api.dns.hostinger.com/v1/records"

HEADERS = {
    "Authorization": f"Bearer {HOSTINGER_API_TOKEN}",
    "Content-Type": "application/json"
}

# Example: addDNSRecord("demo.example.com", "1.2.3.4")
def add_dns_record(subdomain: str, public_ip: str) -> bool:
    """
    Adds an A record for subdomain -> public_ip using Hostinger DNS API.
    """
    try:
        parts = subdomain.split(".")
        if len(parts) < 3:
            logging.error("Subdomain must be in the format sub.domain.tld")
            return False

        domain = ".".join(parts[-2:])
        record_name = ".".join(parts[:-2])

        payload = {
            "type": "A",
            "name": record_name,
            "value": public_ip,
            "ttl": 300
        }

        logging.info(f"Adding A record for {subdomain} → {public_ip}")
        response = requests.post(
            f"https://api.hostinger.com/v1/domains/{domain}/records",
            headers=HEADERS,
            json=payload
        )

        if response.status_code == 201:
            logging.info(f"DNS record created for {subdomain}")
            return True
        else:
            logging.error(f"Failed to create DNS record: {response.text}")
            return False
    except Exception as e:
        logging.error(f"Exception in add_dns_record: {e}")
        return False
