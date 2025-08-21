Cloudflare: How to get CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID

This document explains how to obtain the Cloudflare API token and Zone ID for a domain (example: `abc.com`) and how to store them in your `.env`.

1) CLOUDFLARE_API_TOKEN (recommended: scoped token)

- Log in to Cloudflare: https://dash.cloudflare.com/
- Click your profile (top-right) → "My Profile" → "API Tokens"
- Click "Create Token"
- Use the template: "Edit zone DNS" (or "Edit DNS" template)
- In "Zone Resources" choose "Specific zone" and select your domain `abc.com`
- In "Permissions" the template will grant `Zone:DNS:Edit` (this is sufficient to create/remove A records)
- Continue to summary and create the token
- COPY the token immediately and store it securely (you will not see it again in the dashboard)

Example (.env):

CLOUDFLARE_API_TOKEN=jhkkkjkjnkjn ...your_token_here

Security note: use a scoped token restricted to the specific zone and only DNS edit permissions. Do not store tokens in public repositories.

2) CLOUDFLARE_ZONE_ID

- In Cloudflare Dashboard, open the site `abc.com`
- On the right side of the Overview page you will see the "API" panel containing the "Zone ID"
- COPY the Zone ID value

Example (.env):

CLOUDFLARE_ZONE_ID=3fa85f64-3fc-2c963f66afa6

3) Test the token (optional)

Use curl to test listing DNS records (replace placeholders):

```sh
curl -X GET "https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records" \
  -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  -H "Content-Type: application/json"
```

A successful response returns JSON with `success: true` and `result` list.

4) How to use in this project

- Put the values in the project's `.env` (or the worker `.env`) used by the host worker and server.

Example `.env` entries:

CLOUDFLARE_API_TOKEN=eyJhbUzI1...your_token_here
CLOUDFLARE_ZONE_ID=3fa85f3f66afa6

5) Troubleshooting

- `401 Unauthorized` from Cloudflare: token incorrect or missing permissions. Re-create token with `Zone:DNS:Edit` for the specific zone.
- `404` zone not found: verify Zone ID and that token has access to that zone.

6) Best practices

- Use scoped tokens (not global API keys) with minimal permissions
- Store tokens in environment variables or a secrets manager (AWS Secrets Manager, Vault) — do not commit to Git
- Rotate tokens periodically

If you want, I can add a small script in `autoship-scripts/` that validates the token and zone id and prints the current A records for `abc.com`.
