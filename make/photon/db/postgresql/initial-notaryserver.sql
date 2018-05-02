CREATE DATABASE notaryserver;
CREATE USER server;
alter user server with encrypted password 'password';
GRANT ALL PRIVILEGES ON DATABASE notaryserver TO server;