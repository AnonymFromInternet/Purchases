### This is the repo for the App, which allows users to make online purchases of some sort of items or subscriptions via Stripe payment service.
### Also has an admin backend system for managing all the purchases, subscriptions and users, reset email and password.
###
#### Can be started with the commands from Makefile:
#### for example:
#### $ make start
#### and so on...

#### Variable STRIPE_SECRET_KEY in Makefile should be initialized for the correct work.
#### Also you should have a mailtrap account, and set values in main files for flags username and password, which will be given from mailtrap.
#### Uses:
[Postgres](https://www.postgresql.org/)

[Stripe-go](https://github.com/stripe/stripe-go)

[Jackc/pgx](https://github.com/jackc/pgx) as a sql driver

[Session Manager](https://github.com/alexedwards/scs)

[Mailtrap](https://mailtrap.io/)