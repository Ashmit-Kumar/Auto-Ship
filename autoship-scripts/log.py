import logging
import logging.handlers
import os

# Use a named logger to avoid affecting global logging
logger = logging.getLogger("autoship")
logger.setLevel(logging.INFO)

# Remove existing handlers to prevent duplicates
logger.handlers.clear()

# Configure console handler
console_handler = logging.StreamHandler()
console_handler.setFormatter(logging.Formatter(
    '{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "request_id": "%(request_id)s", "error": "%(error)s"}'
))
logger.addHandler(console_handler)

# Configure file handler with rotation
log_file = os.getenv("LOG_FILE", "/var/log/autoship.log")
os.makedirs(os.path.dirname(log_file), exist_ok=True)
file_handler = logging.handlers.RotatingFileHandler(
    log_file,
    maxBytes=10*1024*1024,  # 10 MB
    backupCount=5
)
file_handler.setFormatter(logging.Formatter(
    '{"time": "%(asctime)s", "level": "%(levelname)s", "message": "%(message)s", "subdomain": "%(subdomain)s", "request_id": "%(request_id)s", "error": "%(error)s"}'
))
logger.addHandler(file_handler)