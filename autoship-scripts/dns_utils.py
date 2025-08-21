import requests
import logging
from config import CLOUDFLARE_API_TOKEN, CLOUDFLARE_ZONE_ID

CF_API_BASE = f"https://api.cloudflare.com/client/v4/zones/{CLOUDFLARE_ZONE_ID}/dns_records"

def add_dns_record(subdomain, ip_address):
    try:
        headers = {
            "Authorization": f"Bearer {CLOUDFLARE_API_TOKEN}",
            "Content-Type": "application/json"
        }

        # Extract the subdomain part (e.g., "test" from "test.example.com")
        record_name = subdomain

        # Check if record already exists (optional but cleaner)
        existing = requests.get(CF_API_BASE, headers=headers, params={"type": "A", "name": record_name}).json()
        for rec in existing.get("result", []):
            # If record exists, delete it to avoid conflicts
            record_id = rec["id"]
            del_resp = requests.delete(f"{CF_API_BASE}/{record_id}", headers=headers)
            if del_resp.status_code not in (200, 204):
                logging.warning(f"Could not delete existing record: {del_resp.text}")

        # Add new A record
        payload = {
            "type": "A",
            "name": record_name,
            "content": ip_address,
            "ttl": 300,
            "proxied": False  # Set to True if you want Cloudflare proxy
        }

        response = requests.post(CF_API_BASE, headers=headers, json=payload)
        if response.status_code in (200, 201):
            logging.info(f"Cloudflare DNS record added: {subdomain} -> {ip_address}")
            return True
        else:
            logging.error(f"Cloudflare DNS add failed: {response.status_code}, {response.text}")
            return False

    except Exception as e:
        logging.error(f"[Cloudflare] Exception adding DNS: {e}")
        return False
