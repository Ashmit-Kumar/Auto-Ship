import subprocess
import logging
import time
import os

CERTBOT_BASE = "/var/lib/autoship/certbot"
CERTBOT_DIRS = {
    "config": os.path.join(CERTBOT_BASE, "config"),
    "work": os.path.join(CERTBOT_BASE, "work"),
    "logs": os.path.join(CERTBOT_BASE, "logs"),
}

CERTBOT_EMAIL = os.getenv("CERTBOT_EMAIL", "2k22.cse.2211344@gmail.com")
CERTBOT_STAGING = os.getenv("CERTBOT_STAGING", "false").lower() in ("1", "true", "yes")
WEBROOT_BASE = os.getenv("CERTBOT_WEBROOT_BASE", "/var/www/autoship")
RSA_KEY_SIZE = int(os.getenv("CERTBOT_RSA_KEY_SIZE", "2048"))


def _is_root():
    try:
        return os.geteuid() == 0
    except AttributeError:
        return False


def _run_cmd(cmd):
    logging.debug(f"Running command: {' '.join(cmd)}")
    return subprocess.run(cmd, capture_output=True, text=True, check=True)


def _ensure_certbot_dirs():
    for path in CERTBOT_DIRS.values():
        os.makedirs(path, exist_ok=True)


def _ensure_webroot(subdomain):
    webroot = os.path.join(WEBROOT_BASE, subdomain)
    challenge = os.path.join(webroot, ".well-known", "acme-challenge")
    os.makedirs(challenge, exist_ok=True)
    return webroot


def generate_ssl_webroot(subdomain, force_renew=False, retries=2, delay=10):
    """Attempt to obtain certificate using webroot plugin."""
    _ensure_certbot_dirs()
    webroot = _ensure_webroot(subdomain)
    base_cmd = [
        "certbot", "certonly", "--webroot",
        "-w", webroot,
        "-d", subdomain,
        "--non-interactive", "--agree-tos",
        "-m", CERTBOT_EMAIL,
        "--no-eff-email",
        "--rsa-key-size", str(RSA_KEY_SIZE),
        "--config-dir", CERTBOT_DIRS["config"],
        "--work-dir", CERTBOT_DIRS["work"],
        "--logs-dir", CERTBOT_DIRS["logs"],
    ]
    if CERTBOT_STAGING:
        base_cmd.insert(1, "--staging")
    if force_renew:
        base_cmd.append("--force-renewal")

    for attempt in range(1, retries + 1):
        try:
            _run_cmd(base_cmd)
            logging.info(f"Webroot certbot succeeded for {subdomain}")
            return True
        except subprocess.CalledProcessError as e:
            logging.error(f"Webroot attempt {attempt}/{retries} failed: {e.stderr}")
            if attempt < retries:
                time.sleep(delay)
    return False


def generate_ssl(subdomain, retries=3, delay=20, prefer_webroot_on_failure=True):
    """
    Generate SSL cert using certbot. Tries the nginx plugin first, then falls back to webroot.
    - Removes `sudo` because service should run as root or be granted rights via systemd.
    - Uses configurable RSA key size and staging option via env.
    """
    logging.info(f"Issuing SSL for: {subdomain}")

    if not _is_root():
        logging.warning("Not running as root — certbot may fail unless granted privileges.")

    _ensure_certbot_dirs()

    nginx_cmd = [
        "certbot", "--nginx",
        "-d", subdomain,
        "--non-interactive", "--agree-tos",
        "-m", CERTBOT_EMAIL,
        "--no-eff-email",
        "--rsa-key-size", str(RSA_KEY_SIZE),
        "--config-dir", CERTBOT_DIRS["config"],
        "--work-dir", CERTBOT_DIRS["work"],
        "--logs-dir", CERTBOT_DIRS["logs"],
    ]
    if CERTBOT_STAGING:
        nginx_cmd.insert(1, "--staging")

    for attempt in range(1, retries + 1):
        try:
            result = _run_cmd(nginx_cmd)
            logging.info(f"Certbot nginx plugin succeeded for {subdomain}: {result.stdout}")
            return True
        except subprocess.CalledProcessError as e:
            logging.error(
                f"Certbot nginx attempt {attempt}/{retries} failed for {subdomain}\n"
                f"Exit code: {e.returncode}\nSTDOUT: {e.stdout}\nSTDERR: {e.stderr}"
            )
            # If nginx plugin misconfiguration likely (nginx test failed), try webroot fallback
            if attempt < retries:
                logging.info(f"Waiting {delay}s before retrying certbot nginx plugin...")
                time.sleep(delay)

    logging.warning(f"All {retries} nginx attempts failed for {subdomain}.")

    if prefer_webroot_on_failure:
        logging.info("Falling back to webroot plugin for cert issuance.")
        # force renewal to replace weak keys if present
        success = generate_ssl_webroot(subdomain, force_renew=True, retries=2, delay=10)
        if success:
            logging.info(f"Webroot fallback succeeded for {subdomain}")
            return True
        else:
            logging.error(f"Webroot fallback also failed for {subdomain}")
            return False

    return False
