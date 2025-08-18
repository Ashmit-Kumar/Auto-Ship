Run locally (development)

1. Prepare .env from .env.example

2. Start MongoDB (or use Atlas). For local dev, run:

```sh
docker run -d -p 27017:27017 --name autoship-mongo mongo:6
```

3. Run Go server locally (bind /var/lib/autoship/deploy for worker communication)

```shn# from autoship-server folder
cp .env.example .env
# optional: build binary
go build -o server ./cmd/server/main.go
# run binary with shared path (create folder first)
mkdir -p /var/lib/autoship/deploy
./server
```

4. Run Python worker on host (recommended)

```sh
cd ../autoship-scripts
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
# run worker (make sure /var/lib/autoship/deploy is writable)
python main.py
```

5. Test: POST repository via /api/projects endpoint and monitor worker logs and response file.
