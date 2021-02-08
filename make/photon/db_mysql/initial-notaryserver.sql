CREATE DATABASE notaryserver;
CREATE USER server;
alter user server IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON notaryserver.* TO server;