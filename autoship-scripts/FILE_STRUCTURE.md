autoship-host/
│
├── host_handler.py          ← Main entry
├── nginx_utils.py           ← Write NGINX confs
├── ssl_utils.py             ← Certbot logic
├── dns_utils.py             ← Hostinger DNS API
├── response_utils.py        ← Send status back to Go app
├── processed.json           ← (Optional) store processed req IDs
└── deploy-requests.json     ← Mounted from container (Go writes here)
