# External Postgres Database

In some scenarios, you might want to leverage an external Postgres database.
For example, if you are deploying to a Kubernetes cluster that is running on
AWS, you could leverage the AWS managed Postgres offering via their RDS service.

If you want to leverage an external Postgres database, you will need to run the
following SQL script:

```sql
-- Create required databases
CREATE DATABASE notaryserver;
CREATE DATABASE notarysigner;
CREATE DATABASE registry ENCODING 'UTF8';
CREATE DATABASE clair;

-- Create harbor user
-- The helm chart limits us to a single user for all databases
CREATE USER harbor;
ALTER USER harbor WITH ENCRYPTED PASSWORD 'change-this-password';

-- Grant the user access to the DBs
GRANT ALL PRIVILEGES ON DATABASE notaryserver TO harbor;
GRANT ALL PRIVILEGES ON DATABASE notarysigner TO harbor;
GRANT ALL PRIVILEGES ON DATABASE registry TO harbor;
GRANT ALL PRIVILEGES ON DATABASE clair to clair;
```