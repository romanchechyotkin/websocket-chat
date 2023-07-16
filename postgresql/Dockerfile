FROM postgres:15.3

ENV POSTGRES_DB chat
ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD 5432

COPY psql_dump.sql /docker-entrypoint-initdb.d/