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

def generate_ssl(subdomain, retries=3, delay=20):
    """
    Generate SSL cert using certbot with nginx plugin.
    Uses custom writable dirs for config, work, and logs.
    Retries in case DNS has not propagated yet.
    """
    logging.info(f"Issuing SSL for: {subdomain}")

    # Ensure custom certbot dirs exist
    for path in CERTBOT_DIRS.values():
        os.makedirs(path, exist_ok=True)

    for attempt in range(1, retries + 1):
        try:
            cmd = [
                "sudo", "certbot", "--nginx",
                "-d", subdomain,
                "--non-interactive", "--agree-tos",
                "-m", "2k22.cse.2211344@gmail.com",
                "--no-eff-email",
                "--config-dir", CERTBOT_DIRS["config"],
                "--work-dir", CERTBOT_DIRS["work"],
                "--logs-dir", CERTBOT_DIRS["logs"],
            ]
            logging.debug(f"Running command: {' '.join(cmd)}")

            result = subprocess.run(
                cmd, capture_output=True, text=True, check=True
            )
            logging.info(f"Certbot succeeded for {subdomain}: {result.stdout}")
            return True

        except subprocess.CalledProcessError as e:
            logging.error(
                f"Certbot attempt {attempt}/{retries} failed for {subdomain}\n"
                f"Exit code: {e.returncode}\n"
                f"STDOUT: {e.stdout}\n"
                f"STDERR: {e.stderr}"
            )
            if attempt < retries:
                logging.info(f"Waiting {delay}s before retrying certbot...")
                time.sleep(delay)

    logging.error(f"All {retries} attempts failed for {subdomain}.")
    return False
