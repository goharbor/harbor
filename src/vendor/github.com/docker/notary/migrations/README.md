# Database Migrations

This directory contains database migrations for the server and signer. They
are being managed using [this tool](https://github.com/mattes/migrate).
Within each of the server and signer directories are directories for different
database backends. Notary server and signer use GORM and are therefore 
capable of running on a number of different databases, however migrations
may contain syntax specific to one backend.
