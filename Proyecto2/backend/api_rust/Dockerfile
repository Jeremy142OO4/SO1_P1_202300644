FROM rust:slim AS build
WORKDIR /app
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app/target/release/api_rust /app/api_rust
ENV BIND=0.0.0.0:8080
ENV GO_API_URL=http://go-deploy1-svc:8080
EXPOSE 8080
CMD ["/app/api_rust"]