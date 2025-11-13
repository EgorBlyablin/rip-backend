# Stage 1: dependencies
FROM node:25-alpine3.22 AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# Stage 2: development container
FROM node:25-alpine3.22 AS dev
WORKDIR /app

# Copy node_modules from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source code separately (faster rebuilds)
COPY . .

EXPOSE 3000

CMD ["npm", "run", "dev"]
