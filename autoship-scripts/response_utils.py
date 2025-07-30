import json
import fcntl
import os

RESPONSE_FILE = "response.json"

def write_response(entry):
    responses = []
    if os.path.exists(RESPONSE_FILE):
        with open(RESPONSE_FILE, "r") as f:
            try:
                fcntl.flock(f, fcntl.LOCK_SH)
                responses = json.load(f)
            except json.JSONDecodeError:
                responses = []
            finally:
                fcntl.flock(f, fcntl.LOCK_UN)

    responses.append(entry)
    with open(RESPONSE_FILE, "w") as f:
        fcntl.flock(f, fcntl.LOCK_EX)
        json.dump(responses[-1000:], f, indent=2)
        fcntl.flock(f, fcntl.LOCK_UN)
