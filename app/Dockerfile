ARG PUBLIC_API_BASE_URL=https://dompet.rasasaufar.site/api

FROM node:22-alpine AS builder
WORKDIR /app
ARG PUBLIC_API_BASE_URL
ENV PUBLIC_API_BASE_URL=$PUBLIC_API_BASE_URL
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npx vite build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules
ENV PORT=3003
ENV HOST=0.0.0.0
EXPOSE 3003
CMD ["node", "build/index.js"]
