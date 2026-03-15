FROM node:24-trixie-slim AS deps
WORKDIR /app

COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

FROM node:24-trixie-slim AS builder
WORKDIR /app

ARG NEXT_PUBLIC_API_BASE_URL=http://api.localhost:8080
ARG NEXT_PUBLIC_APP_BASE_URL=http://collabsphere.localhost:3002
ARG NEXT_INTERNAL_API_BASE_URL=http://api:8080

ENV NEXT_PUBLIC_API_BASE_URL=${NEXT_PUBLIC_API_BASE_URL}
ENV NEXT_PUBLIC_APP_BASE_URL=${NEXT_PUBLIC_APP_BASE_URL}
ENV NEXT_INTERNAL_API_BASE_URL=${NEXT_INTERNAL_API_BASE_URL}

COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm run build

FROM node:24-trixie-slim AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
