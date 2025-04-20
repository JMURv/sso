FROM node:22.4.1-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm install

FROM node:22.4.1-alpine AS builder
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN npm run build

FROM node:22.4.1-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=4000
ENV HOSTNAME='0.0.0.0'

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/.env.local ./.env

EXPOSE 4000

CMD ["node", "server.js"]
