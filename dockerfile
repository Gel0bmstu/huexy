FROM ubuntu:18.04 AS release

MAINTAINER gel0

ENV PGVER 10

USER root

# Install postgres
RUN apt-get update
RUN apt-get install -y postgresql-$PGVER
RUN apt-get install -y curl gnupg2

# Postgres tune
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

RUN apt-get install -y wget

# Create database for repeater
USER postgres

RUN /etc/init.d/postgresql start &&\
    psql -c "CREATE USER gel0 WITH SUPERUSER PASSWORD '1337';" &&\
    createdb -O gel0 proxy &&\
    psql -c "GRANT ALL ON DATABASE proxy TO gel0;" &&\
    /etc/init.d/postgresql stop

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

# Installing go
RUN wget https://dl.google.com/go/go1.12.7.linux-amd64.tar.gz
RUN tar -xvf go1.12.7.linux-amd64.tar.gz
RUN mv go /usr/local
RUN apt-get install -y git

# Set env value for project building
ENV GOPATH /opt/go

ENV PATH $PATH:/usr/local/go/bin

RUN go get github.com/jackc/pgx

ADD / /

# Поiхали
EXPOSE 5000

CMD service postgresql start && go run main.go