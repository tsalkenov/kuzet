FROM debian:latest

RUN ls -la

COPY target/release/scanner-service .

RUN ls -la

EXPOSE 50051

CMD ["/scanner-service"]