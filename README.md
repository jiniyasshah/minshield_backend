# MiniShield Backend

MiniShield is a self-hosted Web Application Firewall (WAF) and DNS management platform. The backend stack includes an HTTP/HTTPS gateway, an ML-based request scorer, PowerDNS with a MariaDB backend, rate limiting / DDoS protection, and transactional email via Brevo SMTP.

## Architecture Overview

| Component                | Description                                                              |
| ------------------------ | ------------------------------------------------------------------------ |
| **Gateway**              | Reverse proxy handling HTTP (80) / HTTPS (443) traffic                   |
| **ML Scorer**            | Machine-learning service that scores incoming requests (`waf_model.pkl`) |
| **PowerDNS**             | Authoritative DNS server (`pdns-auth-49`) with REST API                  |
| **MariaDB**              | Database backend for PowerDNS zone data                                  |
| **MongoDB Atlas**        | Primary application datastore                                            |
| **Brevo SMTP**           | Outbound transactional email                                             |
| **Cloudflare Turnstile** | Bot protection / CAPTCHA                                                 |

## Prerequisites

- A VPS (e.g., DigitalOcean Droplet) running Ubuntu 22.04+ with a public IPv4 address
- A registered domain with nameservers pointed to your DNS provider (e.g., DigitalOcean)
- [Docker Engine](https://docs.docker.com/engine/install/) and the Docker Compose plugin
- A MongoDB Atlas cluster
- A [Brevo](https://www.brevo.com/) account (SMTP credentials + domain verification)
- A [Cloudflare Turnstile](https://www.cloudflare.com/products/turnstile/) site key & secret

## 1. Server Preparation

Update the system and install common utilities:

```bash
apt update
apt upgrade -y
apt full-upgrade -y
apt autoremove -y
apt autoclean

apt install -y curl wget git unzip zip htop tree ncdu tmux screen \
  net-tools dnsutils jq rsync ca-certificates gnupg
```

## 2. Free Port 53 (Required for PowerDNS)

Ubuntu's `systemd-resolved` binds to port 53 by default, which conflicts with PowerDNS. Disable its stub listener:

```bash
sudo nano /etc/systemd/resolved.conf
```

Change:

```ini
#DNSStubListener=yes
```

to:

```ini
DNSStubListener=no
```

Then restart the resolver:

```bash
sudo systemctl restart systemd-resolved
```

## 3. Clone the Repository

```bash
git clone https://github.com/jiniyasshah/minishield_backend.git
cd minishield_backend
```

## 4. Configure Environment Variables

Create your `.env` file from the provided example:

```bash
cp .env.example .env
nano .env
```

Fill in all required values (MongoDB URI, database passwords, SMTP credentials, Turnstile keys, your VPS public IP, etc.). See [`.env.example`](./.env.example) for the full list of variables.

> **Never commit your `.env` file.** Ensure it is listed in `.gitignore`.

## 5. DNS Configuration

After pointing your domain's nameservers to your DNS provider (e.g., DigitalOcean), create the following records for your domain (replace `example.com` with your domain):

| Type    | Hostname           | Value                                       | TTL  |
| ------- | ------------------ | ------------------------------------------- | ---- |
| `A`     | `*.ns.example.com` | `<Your VPS IP>`                             | 3600 |
| `A`     | `api.example.com`  | `<Your VPS IP>`                             | 3600 |
| `CNAME` | `www.example.com`  | `<Your Vercel CNAME>`                       | 3600 |
| `TXT`   | `example.com`      | `brevo-code:<your-brevo-verification-code>` | 3600 |

Additionally, add all DNS records (DKIM, SPF, etc.) provided by Brevo during domain verification to enable email deliverability.

## 6. Build and Run

```bash
docker compose build
docker compose up
```

To run in the background:

```bash
docker compose up -d
```

View logs:

```bash
docker compose logs -f
```

Stop the stack:

```bash
docker compose down
```

## Environment Variables Reference

| Variable                                   | Description                                        |
| ------------------------------------------ | -------------------------------------------------- |
| `GATEWAY_HTTP_PORT` / `GATEWAY_HTTPS_PORT` | Gateway listening ports (default `80` / `443`)     |
| `MONGO_URI`                                | MongoDB Atlas connection string                    |
| `ML_SCORER_URL` / `ML_PORT`                | Internal URL and port of the ML scoring service    |
| `MODEL_FILE_PATH`                          | Path to the trained WAF model inside the container |
| `DNS_DB_*`                                 | MariaDB credentials for the PowerDNS backend       |
| `FRONTEND_URL`                             | Public URL of the frontend application             |
| `WAF_PUBLIC_IP`                            | Public IPv4 address of your VPS                    |
| `DDOS_LIMIT` / `RATE_LIMIT`                | Request rate-limiting thresholds                   |
| `SMTP_USER` / `SMTP_PASS`                  | Brevo SMTP credentials                             |
| `RECAPTCHA_SITE_KEY` / `RECAPTCHA_SECRET`  | Cloudflare Turnstile keys                          |
| `PDNS_*`                                   | PowerDNS image, ports, and config path             |
| `MARIADB_IMAGE`                            | MariaDB Docker image tag                           |
| `SCHEMA_HOST_PATH`                         | Path to the PowerDNS database schema               |

## Troubleshooting

- **PowerDNS fails to start on port 53** — confirm `DNSStubListener=no` is set and `systemd-resolved` was restarted (`sudo ss -tulnp | grep :53` should show no listener before startup).
- **Emails not sending** — verify the Brevo TXT/DKIM/SPF records have propagated (`dig TXT example.com`) and the SMTP key is valid.
- **DNS records not resolving** — allow up to 24–48 hours for nameserver propagation after switching to your DNS provider.

## License

This project is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.

## Contributing

Contributions are welcome. Please open an issue to discuss significant changes before submitting a pull request.
