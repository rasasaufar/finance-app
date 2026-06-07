# 💰 Finance App

> A clean, mobile-first personal finance dashboard for smarter money decisions.

![Svelte](https://img.shields.io/badge/Svelte-FF3E00?style=for-the-badge&logo=svelte&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)
![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

Finance App is a clean, mobile-first personal finance dashboard for tracking money, budgets, and financial insights. It combines a modern SvelteKit interface with a Go backend to help manage daily financial activity, long-term savings goals, and practical reporting in one focused application.

## ✨ Features

### Personal Finance Management

Track income and expenses with category, amount, date, and note details. Finance App keeps everyday money records organized, searchable, and easy to review from a mobile-first interface.

### Multi-Currency Support

Work with currency-aware money values across transactions, budgets, and reports. The experience is optimized for Indonesian Rupiah formatting while keeping the financial model flexible for multi-currency usage.

### Budget Tracking

Create daily, weekly, or monthly budget rules by category and monitor spending against those limits. Overspending alerts help surface budget risks early so adjustments can be made before they become larger problems.

### Financial Analytics

Review dashboard summaries and monthly reports covering income, expenses, net balance, and spending by category. These insights make it easier to understand cash flow patterns and spot where money is going.

### Account & Profile Management

Manage secure login, profile details, avatar uploads, and administrative account controls. The account area is designed to keep personal finance data tied to a clear user profile and role-based access model.

### Wedding Savings

Plan and track wedding savings with a dedicated goal-focused savings tracker. This feature helps monitor progress toward wedding-related targets alongside the rest of the personal finance dashboard.

## 🧰 Tech Stack

| Layer | Technology |
| --- | --- |
| Frontend | SvelteKit, Svelte, TypeScript, CSS |
| Backend | Go, chi router |
| Database | PostgreSQL |
| Deployment | Docker, Docker Compose |

### Docker Architecture

| Service | Role |
| --- | --- |
| Frontend | SvelteKit app served via Nginx reverse proxy |
| API | Go backend served via Nginx reverse proxy |
| PostgreSQL | Database (internal only) |

## 📋 Prerequisites

Before running the project, make sure you have the following installed:

- Git
- Node.js
- Go
- Docker
- Docker Compose

## 🚀 Installation & Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd finance-app
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Configure your environment variables in `.env` (see `.env.example` for reference).

Required environment variables:

```text
DATABASE_URL
JWT_SECRET
ADMIN_PASSWORD
PUBLIC_API_BASE_URL
```

### 3. Run with Docker Compose

```bash
docker compose up --build
```

Access the app at the configured frontend URL.

### 4. Run Locally

Install frontend dependencies:

```bash
cd app
npm install
```

Start the frontend:

```bash
npm run dev
```

Start the backend in a separate terminal:

```bash
cd api
go run ./cmd/api
```

## 📁 Project Structure

```text
finance-app/
├── api/
│   ├── cmd/api/
│   ├── internal/handler/
│   ├── internal/middleware/
│   ├── internal/store/
│   ├── internal/types/
│   ├── internal/validate/
│   ├── Dockerfile
│   └── go.mod
├── app/
│   ├── src/lib/
│   ├── src/routes/
│   ├── static/
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
└── README.md
```

## 🔌 API Overview

The API provides RESTful endpoints for managing transactions, categories, budget rules, salary records, and wedding savings. Protected endpoints require JWT authentication. Full API documentation is available in `docs/API.md`.

## 🖼️ Screenshots

Screenshots will be added here as the interface evolves.

```text
Placeholder: Dashboard screenshot
Placeholder: Transactions screenshot
Placeholder: Budget tracking screenshot
```

## 🤝 Contributing

Contributions are welcome. To contribute, create a feature branch, keep changes focused, follow the existing project structure, and open a pull request with a clear summary of the update.

Please avoid committing secrets, credentials, generated build output, or environment-specific configuration.

## 📄 License

This project is licensed under the MIT License.
