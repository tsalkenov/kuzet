FROM rustlang/rust:nightly-alpine AS rust

WORKDIR /app

RUN apk add --no-cache musl-dev bash curl
RUN curl -L --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install-from-binstall-release.sh | bash
RUN /usr/local/cargo/bin/cargo-binstall --no-confirm cargo-chef

FROM rust as plan

COPY src src/
COPY Cargo.toml .

RUN cargo chef prepare --recipe-path recipe.json

FROM rust AS build
COPY --from=plan /app/recipe.json recipe.json

RUN cargo chef cook --release --recipe-path recipe.json

COPY src src/
COPY Cargo.toml .

RUN cargo build --release && \
    objcopy --compress-debug-sections target/release/scanner-service ./scanner-service

FROM alpine as deploy

WORKDIR /deploy

COPY --from=build /app/scanner-service ./
