# Use the official Rust image as the base image
FROM rust:slim as builder

# Create a new empty shell project
RUN USER=root cargo new --bin scanner-service
WORKDIR /scanner-service

# Copy your manifests
COPY ./Cargo.toml ./Cargo.toml
COPY ./Cargo.lock ./Cargo.lock
COPY ./target ./target
COPY ./src ./src
COPY ./build.rs ./build.rs

# This step will cache your dependencies
RUN cargo build --release
RUN rm src/*.rs

# Now that the dependencies are built, copy your source code

# Build your application with the dependencies cached
RUN rm ./target/release/deps/scanner_service*
RUN cargo build --release

# Use a Debian image with the required GLIBC version to run the scanner service
FROM debian:bookworm-slim
COPY --from=builder /scanner-service/target/release/scanner-service /usr/local/bin/scanner-service

EXPOSE 50051

# Command to run the binary
CMD ["scanner-service"]
