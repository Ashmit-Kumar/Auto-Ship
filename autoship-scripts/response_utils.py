import json
import fcntl
import os
from pathlib import Path
from tempfile import NamedTemporaryFile

# Write responses to host shared path
RESPONSE_FILE = "/var/lib/autoship/deploy/deploy-responses.json"
MAX_RESPONSES = 5000

def write_response(entry: dict):
    Path(os.path.dirname(RESPONSE_FILE)).mkdir(parents=True, exist_ok=True)
    # Read existing
    try:
        with open(RESPONSE_FILE, "r") as f:
            try:
                data = json.load(f)
                if not isinstance(data, list):
                    data = []
            except Exception:
                data = []
    except FileNotFoundError:
        data = []

    # Append new entry and cap
    data.append(entry)
    data = data[-MAX_RESPONSES:]

    # Atomic write via NamedTemporaryFile + os.replace
    dirpath = os.path.dirname(RESPONSE_FILE)
    with NamedTemporaryFile("w", dir=dirpath, delete=False) as tf:
        json.dump(data, tf)
        tf.flush()
        os.fsync(tf.fileno())
        tmpname = tf.name
    os.replace(tmpname, RESPONSE_FILE)
