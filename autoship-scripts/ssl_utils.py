import subprocess
import logging
import os
import re
import dns.resolver
from tenacity import retry, stop_after_attempt, wait_fixed

# Configure structured logging
logging.basicConfig(level=logging.INFO, format='{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "error": "%(error)s"}')

# Configurable admin email
ADMIN_EMAIL = os.getenv("CERTBOT_ADMIN_EMAIL", "admin@example.com")

def validate_subdomain(subdomain: str) -> bool:
    """Validate subdomain to prevent injection attacks."""
    if not isinstance(subdomain, str):
        logging.error("Subdomain must be a string", extra={"subdomain": str(subdomain), "error": "Invalid type"})
        return False
    if not re.match(r'^[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9]$', subdomain):
        logging.error("Invalid subdomain format", extra={"subdomain": subdomain, "error": "Invalid characters"})
        return False
    if '..' in subdomain or '/' in subdomain or ';' in subdomain:
        logging.error("Subdomain contains unsafe characters", extra={"subdomain": subdomain, "error": "Injection risk"})
        return False
    return True

def check_certificate_exists(subdomain: str) -> bool:
    """Check if a valid certificate exists for the subdomain."""
    cert_path = f"/etc/letsencrypt/live/{subdomain}/fullchain.pem"
    if os.path.exists(cert_path):
        try:
            result = subprocess.run(
                ["openssl", "x509", "-in", cert_path, "-noout", "-checkend", "0"],
                capture_output=True,
                text=True
            )
            if result.returncode == 0:
                logging.info("Valid certificate already exists", extra={"subdomain": subdomain, "error": ""})
                return True
        except subprocess.SubprocessError as e:
            logging.error("Error checking certificate", extra={"subdomain": subdomain, "error": str(e)})
    return False

def check_dns_propagation(subdomain: str, public_ip: str, max_attempts: int = 5, delay: int = 30) -> bool:
    """Check if DNS A record has propagated."""
    for attempt in range(1, max_attempts + 1):
        try:
            answers = dns.resolver.resolve(subdomain, "A")
            for rdata in answers:
                if str(rdata) == public_ip:
                    logging.info(f"DNS propagated for {subdomain}", extra={"subdomain": subdomain, "error": ""})
                    return True
            logging.warning(f"DNS not yet propagated (attempt {attempt}/{max_attempts})", extra={"subdomain": subdomain, "error": ""})
        except Exception as e:
            logging.warning(f"DNS check failed (attempt {attempt}/{max_attempts})", extra={"subdomain": subdomain, "error": str(e)})
        if attempt < max_attempts:
            time.sleep(delay)
    logging.error("DNS propagation failed after max attempts", extra={"subdomain": subdomain, "error": "Timeout"})
    return False

@retry(stop=stop_after_attempt(3), wait=wait_fixed(10))
def generate_ssl(subdomain: str) -> bool:
    """
    Generates SSL certificate for a subdomain using Certbot with NGINX plugin.
    """
    if not validate_subdomain(subdomain):
        raise ValueError(f"Invalid subdomain: {subdomain}")

    if check_certificate_exists(subdomain):
        return True

    try:
        logging.info("Generating SSL certificate", extra={"subdomain": subdomain, "error": ""})
        result = subprocess.run(
            ["certbot", "--nginx", "-d", subdomain, "--non-interactive", "--agree-tos", "-m", ADMIN_EMAIL],
            capture_output=True,
            text=True
        )
        if result.returncode == 0:
            logging.info("SSL generation successful", extra={"subdomain": subdomain, "error": ""})
            return True
        else:
            error_msg = result.stderr
            logging.error("Certbot failed", extra={"subdomain": subdomain, "error": error_msg})
            if "too many requests" in error_msg.lower():
                raise RuntimeError("Certbot rate limit exceeded")
            return False
    except subprocess.SubprocessError as e:
        logging.error("Subprocess error during SSL generation", extra={"subdomain": subdomain, "error": str(e)})
        return False
    except Exception as e:
        logging.error("Unexpected error during SSL generation", extra={"subdomain": subdomain, "error": str(e)})
        return False