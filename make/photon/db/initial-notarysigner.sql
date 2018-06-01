CREATE DATABASE notarysigner;
CREATE USER signer;
alter user signer with encrypted password 'password';
GRANT ALL PRIVILEGES ON DATABASE notarysigner TO signer;