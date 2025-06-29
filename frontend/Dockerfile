FROM node:22.4.1-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm install

FROM node:22.4.1-alpine AS builder
WORKDIR /app

ARG BACKEND_URL
ARG S3_URL
ARG NEXT_PUBLIC_RECAPTCHA_SITE_KEY
ARG NEXT_PUBLIC_RECAPTCHA_SECRET_KEY

ENV BACKEND_URL=$BACKEND_URL
ENV S3_URL=$S3_URL
ENV NEXT_PUBLIC_RECAPTCHA_SITE_KEY=$NEXT_PUBLIC_RECAPTCHA_SITE_KEY
ENV NEXT_PUBLIC_RECAPTCHA_SECRET_KEY=$NEXT_PUBLIC_RECAPTCHA_SECRET_KEY

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

EXPOSE 4000

CMD ["node", "server.js"]
