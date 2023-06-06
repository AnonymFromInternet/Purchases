### This is the repo for the App, which allows users to make online purchases of some sort of items via Stripe payment service.
#### Can be started with the commands from Makefile:
#### $ make start and so on.

#### Variable STRIPE_SECRET_KEY in Makefile should be initialized for the correct work.
#### Uses:
[Postgres](https://www.postgresql.org/)

[Stripe-go](https://github.com/stripe/stripe-go)

[Jackc/pgx](https://github.com/jackc/pgx) as a sql driver

[Session Manager](https://github.com/alexedwards/scs)