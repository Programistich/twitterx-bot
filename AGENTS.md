This is a Telegram Bot what use https://github.com/Programistich/twitterx-api to fetch Twitter user data and send it to Telegram users.

Always check Telegram limits before implementing new features, like max message length, photo count, etc.

Look @infra folder for deployment configs

## Run with Docker

Docker Compose lives in `infra/`.

1) Ensure `infra/.env` exists (you can copy `infra/.env.example` and adjust values).
2) From repo root:
   - Development (local ports): `docker compose -f infra/docker-compose.yml -f infra/docker-compose.override.yml up -d --build`
   - Production (shared proxy network): `docker compose -f infra/docker-compose.yml -f infra/docker-compose.prod.yml up -d --build`
