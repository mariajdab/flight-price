# Flight service

This project is designed to retrieve flight information from three different providers—**Amadeus**, **Sky (Rapid API)**, and **Google Flights (Rapid API)**—given an origin, destination, and date.

## Features
- Fetches flight data from multiple APIs using concurrency.
- Uses **Docker secrets** to securely handle sensitive API keys (an alternative approach could use AWS Secrets Manager).
- Supports HTTPS with self-signed certificates. Automatically works with Let's Encrypt certificates if a real domain is configured. If we don't have a real domain we need to create self signed certificates
- Includes a helper to convert city names to provider-specific codes (e.g., "Paris" becomes `PARI` for Sky API). Currently supports **14 cities** (see `helper.go`).

## Prerequisites
- API keys/secrets for:
    - Amadeus API
    - Sky Rapid API
    - Google Flights Rapid API
- Docker and Docker Compose installed

## Setup
1. Create a `secrets` directory in the project root.
2. Add your API credentials as files with these **exact names**:
    - `amadeus_api_key.txt`
    - `amadeus_api_secret.txt`
    - `sky_rapid_api_key.txt`
    - `google_flight_rapid_api_key.txt`

## Running the Project (Development)
1. Generate self-signed certificates (run in `src/` directory):
   `
   openssl req -x509 -newkey rsa:4096 -keyout cert.key -out cert.pem -days 365 -nodes
`
2. Start the container:
`
docker compose up --build
`

4. Server will run on port 8443 with HTTPS.

3. Authenticate:
`
https://localhost:8443/public/
`

After authentication, you'll be redirected to the flight search interface.

Notes

    Default environment: development (uses self-signed certs)

    To use Let's Encrypt certificates, configure a domain and set the environment to production.
