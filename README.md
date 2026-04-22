# Better — NBA Analytics & Intelligence Platform

Full-stack probabilistic analysis platform for NBA games. Provides statistical analysis, odds comparison, value detection, and manual bankroll tracking.

> ⚠️ **Disclaimer**: This platform is for **informational and educational purposes only**. It does not accept, facilitate, or encourage any form of gambling. Past performance does not guarantee future results.

## Architecture

- **Backend**: Go 1.22+ (gorilla/mux, pgx/v5, go-redis)
- **Frontend**: Next.js 14 (App Router, TypeScript, Tailwind CSS, Recharts)
- **Database**: PostgreSQL 16 + Redis 7
- **Auth**: JWT (HS256) with refresh token rotation
- **Infra**: Docker Compose, GitHub Actions CI/CD

## Quick Start

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your API keys (NBA Stats, The Odds API)

# Start all services
docker compose up --build

# Access
# Frontend: http://localhost:3000
# Backend:  http://localhost:8080/api/v1
```

## Development

### Backend
```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Tests
```bash
# Backend
cd backend && go test ./tests/...

# Frontend
cd frontend && npm test

# E2E
cd e2e && npx playwright test
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | /api/v1/auth/register | No | Register |
| POST | /api/v1/auth/login | No | Login |
| POST | /api/v1/auth/refresh | No | Refresh tokens |
| POST | /api/v1/auth/logout | Yes | Logout |
| GET | /api/v1/auth/me | Yes | Current user |
| GET | /api/v1/analysis/games/today | Yes | Today's games |
| GET | /api/v1/analysis/games/:id | Yes | Game analysis |
| GET/PUT | /api/v1/user/preferences | Yes | User preferences |
| GET | /api/v1/user/export | Yes | LGPD data export |
| DELETE | /api/v1/user/data | Yes | LGPD right to deletion |
| GET/POST | /api/v1/bankroll/entries | Yes | Bankroll entries |
| GET | /api/v1/bankroll/dashboard | Yes | Performance stats |
| POST | /api/v1/admin/scraper/run | Admin | Trigger scrape |

## LGPD Compliance

- Full data export (GET /api/v1/user/export)
- Right to deletion (DELETE /api/v1/user/data)
- Explicit consent at registration
- Audit logging

## License

Private — All rights reserved.
