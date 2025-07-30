import subprocess
import logging

def generate_ssl(subdomain):
    try:
        logging.info(f"Issuing SSL for: {subdomain}")
        subprocess.run(
            ["sudo", "certbot", "--nginx", "-d", subdomain, "--non-interactive", "--agree-tos", "-m", "admin@example.com"],
            check=True
        )
        return True
    except subprocess.CalledProcessError as e:
        logging.error(f"Certbot failed: {e}")
        return False
