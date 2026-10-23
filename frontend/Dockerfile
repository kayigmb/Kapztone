# syntax=docker/dockerfile:1
FROM node:22-alpine AS build
WORKDIR /app
RUN corepack enable
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY . .
# The runtime is Alpine (musl): drop sharp's glibc build that standalone traced in too.
RUN pnpm build && rm -rf .next/standalone/node_modules/.pnpm/@img+sharp-libvips-linux-* \
    .next/standalone/node_modules/.pnpm/@img+sharp-linux-*

# Alpine's own Node (no npm, yarn or corepack) runs Next's standalone server.
FROM alpine:3.22
RUN apk add --no-cache nodejs && adduser -S -H app
WORKDIR /app
ENV NODE_ENV=production HOSTNAME=0.0.0.0 PORT=3000
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
# Next writes optimised images here at run time.
RUN mkdir -p .next/cache && chown app .next/cache
USER app
EXPOSE 3000
CMD ["node", "server.js"]
