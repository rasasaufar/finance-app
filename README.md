# 💰 Finance App

> A clean, mobile-first personal finance dashboard for tracking money, budgets, and financial insights.

![Svelte](https://img.shields.io/badge/Svelte-FF3E00?style=for-the-badge&logo=svelte&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

Finance App is a personal finance management application built with a SvelteKit frontend and a Go API. It helps a single user record transactions, organize categories, monitor budgets, and review financial performance through a simple fintech-style interface.

## ✨ Features

### Personal Finance Management

Track income and expenses with category, amount, date, and note details. The app is designed for everyday personal use with a fast, mobile-friendly workflow.

### Multi-Currency Support

Structure finances around currency-aware money values so transactions, budgets, and analytics can support multiple currencies. The current interface is optimized for Indonesian Rupiah formatting.

### Budget Tracking

Create daily, weekly, or monthly budget rules by category. Budget checks are returned when transactions are created or updated, making it easier to spot overspending early.

### Financial Analytics

Review dashboard summaries and monthly reports, including income, expenses, net balance, spending by category, and budget usage.

### Account & Profile Management

Use a seeded admin account for login, update profile information, manage avatar uploads, and manage accounts through admin-only API routes.

## 🧰 Tech Stack

| Layer | Technology |
| --- | --- |
| Frontend | SvelteKit, Svelte, TypeScript, CSS |
| Backend | Go, chi router |
| Storage | PostgreSQL-backed API storage |
| Deployment | Docker, Docker Compose |

## 📋 Prerequisites

Make sure these tools are installed before running the project:

- [Go](https://go.dev/doc/install) 1.26 or newer
- [Node.js](https://nodejs.org/) 20 or newer
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- PostgreSQL access for the API `DATABASE_URL`

## 🚀 Installation & Setup

### 1. Clone the Repository

```bash
git clone https://github.com/rasasaufar/finance-app.git
cd finance-app
```

### 2. Configure Environment Variables

Create `app/.env`:

```bash
cd app
printf "PUBLIC_API_BASE_URL=http://localhost:8080\n" > .env
```

Configure the API database connection with `DATABASE_URL`:

```env
DATABASE_URL=postgres://finance_user:finance_pass_2026@localhost:5432/finance_app?sslmode=disable
```

The current `docker-compose.yml` expects PostgreSQL to be reachable from the API container through `host.docker.internal`. For local-only Docker usage, set the frontend build argument `PUBLIC_API_BASE_URL` in `docker-compose.yml` to `http://localhost:8080` before building.

### 3. Run with Docker Compose

From the project root:

```bash
docker compose up --build
```

Services:

- Frontend: `http://localhost:3003`
- API: `http://localhost:8080`

Default seeded login:

```text
Email: ***REMOVED***
Password: ***REMOVED***
```

### 4. Run Locally Without Docker

Start the API:

```bash
cd api
DATABASE_URL="postgres://finance_user:finance_pass_2026@localhost:5432/finance_app?sslmode=disable" go run ./cmd/api
```

Start the frontend in another terminal:

```bash
cd app
npm install
npm run dev
```

The frontend development server usually runs at `http://localhost:5173`.

## 📁 Project Structure

```text
finance-app/
├── api/
│   ├── cmd/api/                 # API entry point
│   ├── internal/handler/         # HTTP handlers by domain
│   ├── internal/middleware/      # Auth and admin middleware
│   ├── internal/store/           # Persistence and seed logic
│   ├── internal/types/           # Shared API request/response types
│   ├── internal/validate/        # Request validation helpers
│   ├── Dockerfile
│   └── go.mod
├── app/
│   ├── src/lib/                  # Frontend API, auth, format, and shared types
│   ├── src/routes/               # SvelteKit pages and layouts
│   ├── static/                   # Icons, manifest, service worker assets
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
└── README.md
```

## 🔌 API Endpoints

Base URL:

```text
http://localhost:8080
```

Protected endpoints require:

```http
Authorization: Bearer <token>
```

### Public Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check endpoint |
| `POST` | `/auth/login` | Authenticate user and return an access token |
| `GET` | `/images/*` | Serve uploaded avatar images |

Example login request:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"***REMOVED***","password":"***REMOVED***"}'
```

### Authenticated User

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/me` | Get the current user profile |
| `PUT` | `/me` | Update the current user profile |
| `POST` | `/me/avatar` | Upload a profile avatar |
| `DELETE` | `/me/avatar` | Delete the profile avatar |

### Dashboard & Reports

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/dashboard/summary` | Get balance, salary, expense, and budget summary |
| `GET` | `/reports/monthly` | Get monthly income, expense, net, category, and budget analytics |

### Transactions

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/transactions` | List transactions |
| `POST` | `/transactions` | Create a transaction |
| `PUT` | `/transactions/{id}` | Update a transaction |
| `DELETE` | `/transactions/{id}` | Delete a transaction |

Example transaction payload:

```json
{
  "type": "expense",
  "category": "Makan",
  "amount": 60000,
  "date": "2026-06-03",
  "note": "Lunch"
}
```

### Categories

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/categories` | List categories |
| `POST` | `/categories` | Create a category |
| `PUT` | `/categories/{id}` | Update a category |
| `DELETE` | `/categories/{id}` | Delete a category |

### Budget Rules

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/budget-rules` | List budget rules |
| `POST` | `/budget-rules` | Create a budget rule |
| `PUT` | `/budget-rules/{id}` | Update a budget rule |
| `DELETE` | `/budget-rules/{id}` | Delete a budget rule |

Example budget rule payload:

```json
{
  "category": "Makan",
  "period": "daily",
  "limit": 60000
}
```

### Salary Masters

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/salary-masters` | List monthly salary records |
| `POST` | `/salary-masters` | Create a salary record |
| `PUT` | `/salary-masters/{id}` | Update a salary record |
| `DELETE` | `/salary-masters/{id}` | Delete a salary record |

### Wedding Savings

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/wedding` | Get wedding savings summary |
| `PUT` | `/wedding/config` | Update wedding savings configuration |
| `POST` | `/wedding/deposits` | Create a wedding savings deposit |
| `PUT` | `/wedding/deposits/{id}` | Update a wedding savings deposit |
| `DELETE` | `/wedding/deposits/{id}` | Delete a wedding savings deposit |

### Admin

Admin endpoints require an authenticated user with the `admin` role.

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/admin/accounts` | List user accounts |
| `POST` | `/admin/accounts` | Create a user account |
| `PUT` | `/admin/accounts/{id}` | Update a user account |
| `DELETE` | `/admin/accounts/{id}` | Delete a user account |

## 🖼️ Screenshots

Add screenshots to this section as the UI evolves.

| Dashboard | Transactions | Reports |
| --- | --- | --- |
| `screenshots/dashboard.png` | `screenshots/transactions.png` | `screenshots/reports.png` |

## 🤝 Contributing

Contributions are welcome. To keep changes easy to review:

1. Fork the repository.
2. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. Install dependencies and run checks before submitting:

   ```bash
   cd app
   npm install
   npm run check
   npm run lint
   ```

   ```bash
   cd api
   go test ./...
   ```

4. Commit with a clear message.
5. Open a pull request describing the problem, solution, and any screenshots for UI changes.

## 📄 License

This project is licensed under the MIT License.
