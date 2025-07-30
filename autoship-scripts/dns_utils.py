import requests
import logging
from config import HOSTINGER_API_KEY, HOSTINGER_ZONE_ID

HOSTINGER_API_URL = f"https://api.hostinger.com/v1/zones/{HOSTINGER_ZONE_ID}/records"

def add_dns_record(subdomain, ip_address):
    try:
        headers = {
            "Authorization": f"Bearer {HOSTINGER_API_KEY}",
            "Content-Type": "application/json"
        }

        name = subdomain.split('.')[0]  # Only the subdomain, like "test"
        data = {
            "type": "A",
            "name": name,
            "value": ip_address,
            "ttl": 300
        }

        response = requests.post(HOSTINGER_API_URL, headers=headers, json=data)
        if response.status_code in (200, 201):
            logging.info(f"DNS record added: {subdomain} -> {ip_address}")
            return True
        else:
            logging.error(f"Failed to add DNS: {response.status_code}, {response.text}")
            return False
    except Exception as e:
        logging.error(f"Exception adding DNS record: {e}")
        return False
