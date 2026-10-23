# syntax=docker/dockerfile:1
# The dev chain. Hardhat compiles the contracts once at build time; the image
# keeps only their bytecode and Foundry's anvil (the chain) and cast (deploys).

FROM node:22-alpine AS build
WORKDIR /c
RUN corepack enable
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY hardhat.config.ts tsconfig.json ./
COPY contracts ./contracts
RUN pnpm hardhat compile && mkdir /out && node -e '\
for (const c of ["OrgRegistry","FacilityRegistry","PassportRegistry","CustodyManager","AttestationRegistry"]) \
  require("fs").writeFileSync(`/out/${c}.hex`, require(`./artifacts/contracts/${c}.sol/${c}.json`).bytecode)'

FROM ghcr.io/foundry-rs/foundry:v1.5.1 AS foundry

# anvil and cast need glibc and libgcc; the debug variant adds busybox for chain.sh.
FROM gcr.io/distroless/cc-debian12:debug-nonroot
COPY --from=foundry /usr/local/bin/anvil /usr/local/bin/cast /usr/local/bin/
COPY --from=build /out /contracts
COPY chain.sh /chain.sh
EXPOSE 8545
HEALTHCHECK --interval=2s --retries=60 CMD ["/busybox/test", "-f", "/tmp/deployed"]
ENTRYPOINT ["/busybox/sh", "/chain.sh"]
