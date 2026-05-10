# Project Instructions

This is a personal finance management app.

## Project Structure

- `app/` contains the SvelteKit frontend.
- `api/` contains the Go backend.

## Stack

- Frontend: SvelteKit with TypeScript
- Backend: Go with chi router
- Initial storage: in-memory only
- Auth: login only, no register

## Product Direction

The app is for one personal user only.

The app should help the user manage:
- transactions
- categories
- daily budgets
- weekly budgets
- monthly budgets
- reports

The app must be mobile-friendly first.

Use Indonesian language for UI copy.

Use Indonesian Rupiah formatting for money.

## Default Budget Examples

- Makan: Rp 60.000 per day
- Bensin: Rp 240.000 per month

## Frontend Requirements

- Use responsive mobile-first layout.
- On desktop, use sidebar navigation.
- On mobile, use bottom navigation or compact navigation.
- Keep UI clean, simple, modern, and fintech-like.
- Forms must be easy to use on mobile.
- Main buttons must be easy to tap.
- Do not create a register page.

## Backend Requirements

- Use Go with chi router.
- Keep implementation simple.
- Use in-memory data only for now.
- Do not add database yet.
- Use hardcoded login for now:
  - email: rasas@example.com
  - password: password123
- Return a dummy token after login.

## Run Commands

Run frontend:

```bash
cd app
npm run dev