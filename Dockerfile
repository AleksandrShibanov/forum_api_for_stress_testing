FROM golang:1.17 AS build

ADD . /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/main/main.go

FROM postgres:alpine

USER postgres

COPY --chmod=777 ./db.sql db.sql

ENV POSTGRES_DB=forum
ENV POSTGRES_USER=aleksandr
ENV PGVER=12

RUN chmod 777 /var/lib/postgresql/data &&\
    initdb /var/lib/postgresql/data &&\
    pg_ctl start &&\
    psql -c "CREATE USER aleksandr WITH SUPERUSER PASSWORD 'password';" &&\
    createdb -h localhost -p 5432 -U $POSTGRES_USER $POSTGRES_DB &&\
    psql -U $POSTGRES_USER -d $POSTGRES_DB -f /db.sql &&\
    echo "host all all 0.0.0.0/0 md5" >> /var/lib/postgresql/data/pg_hba.conf &&\
    echo "local all postgres ident" >> /var/lib/postgresql/data/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf &&\
    echo "shared_buffers=256MB" >> /var/lib/postgresql/data/postgresql.conf &&\
    echo "full_page_writes=off" >> /var/lib/postgresql/data/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
EXPOSE 5000

COPY --from=build /app/main .
CMD pg_ctl start && ./main
