CREATE DATABASE notarysigner;
CREATE USER signer;
alter user signer IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON notarysigner.* TO signer;