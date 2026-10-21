FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_BASE_URL=/api
ARG VITE_IDLE_TIMEOUT_MS=900000
ARG VITE_REQUEST_TIMEOUT_MS=20000
ARG VITE_MAX_RETRIES=4

ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_IDLE_TIMEOUT_MS=$VITE_IDLE_TIMEOUT_MS
ENV VITE_REQUEST_TIMEOUT_MS=$VITE_REQUEST_TIMEOUT_MS
ENV VITE_MAX_RETRIES=$VITE_MAX_RETRIES

RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
