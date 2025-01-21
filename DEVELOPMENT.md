# Development

In this section, you can find the instructions needed for local development.

---

## Requirements

To run, test, or develop on your local machine, it's recommended to install
the following requirements. A Docker Compose file is also provided, which
includes all the necessary services.

- `go` - current version: `1.23.4`
- `postgresql` - version `16` or higher
- `pre-commit` - optional, use `pre-commit install` if you plan to contribute!
- `direnv` - optional, recommended, manages your environment variables over `.envrc`
- `docker` - optional, recommended, you can run external/all services with `docker-compose`
- `ngrok` - if you want to receive webhooks locally.
- `golang-migrate` - for database revision management.

If you are on macOS or Linux and using [homebrew][001], you can install with
the following command:

```bash
brew install go postgresql pre-commit direnv golang-migrate
brew install ngrok # optional
```

We highly recommend [Orbstack][002] for macOS instead of `docker` community
edition.

If you plan to add new `rake` tasks, please ensure that `ruby` and `rubocop` are
also installed on your local machine:

```bash
brew install ruby
gem update --system
gem install bundler   # or
gem install --user-install bundler

bundle config set --local path 'ruby-vendor/bundle'
bundle config set --local bin 'bin'
bundle
```

---

## Environment Variables

| Variable | Description | Default |
|:---------|:------------|---------|
| `LISTEN_ADDR` | Server listen address | `":8000"` |
| `LOG_LEVEL` | Logging level, Valid values are: `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"` | `"INFO"` |
| `GITHUB_HMAC_SECRET` | HMAC secret value for GitHub’s webhooks. | `""` |
| `KCP_BROKERS` | Kafka consumer/producer brokers list, comma separated | `"127.0.0.1:9094"` |
| `KP_GITHUB_MESSAGE_QUEUE_SIZE` | Size of default Kafka message producer queue size for GitHub Webhooks | `100` |
| `KC_PARTITION` | Consumer partition number | `0` |
| `KC_TOPIC` | Topic to subscribe | `""` |
| `KC_DIAL_TIMEOUT` | Initial connection timeout used by broker | "`30s`" (seconds) |
| `KC_READ_TIMEOUT` | Response timeout used by broker | "`30s`" (seconds) |
| `KC_WRITE_TIMEOUT` | Transmit timeout used by broker | "`30s`" (seconds) |
| `KC_BACKOFF` | Backoff value for retries | "`2s`" (seconds) |
| `KC_MAX_RETRIES` | Maximum retry | `10` |
| `DATABASE_NAME` | Name of your PostgreSQL database | `""` |
| `DATABASE_URL` | PostgreSQL dsn | `""` |
| `POSTGRES_DB`: PostgreSQL database name for docker-compose | `""` |
| `POSTGRES_PASSWORD`: PostgreSQL password for docker-compose | `""` |
| `DATABASE_URL_DOCKER_TO_HOST`: For testing migrator to connect you local pg | `""` |
| `DATABASE_URL_INFRA`: For postgresql dsn for infra | `""` |
| `NGROK_AUTHTOKEN`: your ngrok auth token | `""` |
| `NGROK_PUBLIC_URL`: your ngrok public url | `""` |
| `NGROK_FORWARD`: your service tcp addr for ngrok | `""` |

Grab your `ngrok` auth token via:

```bash
ngrok config edit
```

You’ll see a `yaml` file:

```yaml
version: "3"
agent:
    authtoken: <your-ngrok-token>
```

Example `.envrc`:

```bash
# for ruby/rake only
export PATH="bin:${PATH}"

# cauldron server
export LISTEN_ADDR=":8000"
export LOG_LEVEL="INFO"
export GITHUB_HMAC_SECRET="<secret>"

# kafka consumer/producer shared values.
export KCP_BROKERS="127.0.0.1:9094"

# kafka producer values.
export KP_GITHUB_MESSAGE_QUEUE_SIZE=100

# kafka github consumer values.
export KC_PARTITION="0"
export KC_DIAL_TIMEOUT="30s"
export KC_READ_TIMEOUT="30s"
export KC_WRITE_TIMEOUT="30s"
export KC_BACKOFF="2s"
export KC_MAX_RETRIES="10"

# database variables, works for local postgresql
export DATABASE_NAME="devchain_webhook"
export DATABASE_URL="postgres://localhost:5432/${DATABASE_NAME}?sslmode=disable&timezone=UTC"

# postgresql runs on infra/docker
export POSTGRES_DB="${DATABASE_NAME}"
export POSTGRES_PASSWORD="gee2Jiquush4eiso"
export DATABASE_URL_INFRA="postgres://postgres:${POSTGRES_PASSWORD}@postgresql_db:5432/${POSTGRES_DB}?sslmode=disable&timezone=UTC"

# used for testing migrator, assuming postgresql is installed from brew.
export DATABASE_URL_DOCKER_TO_HOST="postgres://${USER}@host.docker.internal:5432/${DATABASE_NAME}?sslmode=disable&timezone=UTC"

# for ngrok
export NGROK_AUTHTOKEN="<your-ngrok-token>"
export NGROK_PUBLIC_URL="<your-ngrok-public-url>"
export NGROK_FORWARD="http://cauldron:8000" # must match with LISTEN_ADDR value (port)
```

---

## Clone The Source for Development

```bash
cd /path/to/development
git clone git@github.com:devchain-network/cauldron.git
cd cauldron/
go mod download
```

Install `pre-commit` hooks before you starting development:

```bash
pre-commit install
```

Open a new branch and do your thing!

---

## Rake Tasks

If `ruby` is installed or available on your machine, you can leverage the power
of `rake tasks`:

```bash
rake -T

rake db:init                       # init database
rake db:migrate                    # runs rake db:migrate up (shortcut)
rake db:migrate:down               # run migrate down
rake db:migrate:goto[index]        # go to migration
rake db:migrate:up                 # run migrate up
rake db:psql                       # connect local db with psql
rake db:reset                      # reset database (drop and create)
rake default                       # default task, runs server
rake docker:build:github_consumer  # build github consumer
rake docker:build:migrator         # build migrator
rake docker:build:server           # build server
rake docker:compose:infra:down     # stop the infra with all components
rake docker:compose:infra:up       # run the infra with all components
rake docker:compose:kafka:down     # stop the kafka and kafka-ui only
rake docker:compose:kafka:up       # run the kafka and kafka-ui only
rake docker:run:github_consumer    # run github consumer
rake docker:run:migrator           # run migrator
rake docker:run:server             # run server
rake lint                          # run golang-ci linter
rake rubocop:autofix               # lint ruby and autofix
rake rubocop:lint                  # lint ruby
rake run:kafka:github:consumer     # run kafka github consumer
rake run:server                    # run server
rake test                          # runs tests (shortcut)
rake test:coverage                 # run tests and show coverage
```

You can run tests:

```bash
rake -T "test"

rake test           # runs tests (shortcut)
rake test:coverage  # run tests and show coverage
```

Display all the `db` related tasks:

```bash
rake -T "db:"

rake db:init                 # init database
rake db:migrate              # runs rake db:migrate up (shortcut)
rake db:migrate:down         # run migrate down
rake db:migrate:goto[index]  # go to migration
rake db:migrate:up           # run migrate up
rake db:psql                 # connect local db with psql
rake db:reset                # reset database (drop and create)
```

If you have `postgresql` locally installed, run the following `rake` tasks to
setup your database locally:

```bash
rake db:init      # create db
rake db:migrate   # run migrations
```

`rake db:reset` task is only used for starting from scratch. Drops and created
db for you.

`rake db:migrate:down` applies all the migrations through down, kind of returning
back to initial. If you want to go to specific state, use:

```bash
rake db:migrate:goto[3]    # go back to `000003_github.up.sql`
```

You can connect your local postgres instance with correct search path with
using `rake db:psql`


You can run each service/component separately with opening multiple terminal
tabs.

```bash
rake docker:compose:kafka:up                     # kick kafka + kafka ui in tab 1
docker compose -f docker-compose.kafka.yml up    # or use this w/o rake.

rake                                # kick webhook server in tab 2.
go run cmd/server/main.go           # or use this w/o rake.

rake run:kafka:github:consumer      # kick github consumer in tab 3
go run cmd/githubconsumer/main.go   # or use this w/o rake.

ngrok http --url=<url-url>.ngrok-free.app 8000 # run ngrok in tab 4
```

Set `https://<url-url>.ngrok-free.app` webhook url on your GitHub repo or
GitHub organization. Don’t forget to set your webhook secret as `GITHUB_HMAC_SECRET`
before kicking webhook server (tab 2).

Now, do some action on your GitHub repository and let the service do it’s thing!

### Run Tasks

```bash
rake -T "^run:"

rake run:kafka:github:consumer   # run kafka github consumer
rake run:server                  # run server
```

### Linter Tasks

```bash
rake lint              # runs golang-ci linter
rake rubocop:lint      # lints ruby code
rake rubocop:autofix   # lints ruby code and auto fixes.
```

### Docker Tasks

```bash
rake -T "docker:"

rake docker:build:github_consumer  # build github consumer
rake docker:build:migrator         # build migrator
rake docker:build:server           # build server

rake docker:compose:infra:down     # stop the infra with all components
rake docker:compose:infra:up       # run the infra with all components
rake docker:compose:kafka:down     # stop the kafka and kafka-ui only
rake docker:compose:kafka:up       # run the kafka and kafka-ui only

rake docker:run:github_consumer    # run github consumer
rake docker:run:migrator           # run migrator
rake docker:run:server             # run server
```

- `docker:build:*`: builds images locally, testing purposes.
- `docker:run:*`: runs containers locally, testing purposes.
- `docker:compose:*`: ups or downs whole infrastructure with services.

---

## Infrastructure Diagram

```mermaid
graph TD
  subgraph devchain-network
    kafka["Kafka"]
    kafka-ui["Kafka UI"]
    cauldron["Cauldron"]
    github-consumer["GitHub Consumer"]
    postgresql_db["PostgreSQL DB"]
    migrator["Migrator"]
    ngrok["ngrok"]
  end

  kafka -->|depends_on| kafka-ui
  kafka -->|depends_on| cauldron
  kafka -->|depends_on| github-consumer
  cauldron -->|depends_on| github-consumer
  github-consumer -->|depends_on| postgresql_db
  github-consumer -->|depends_on| migrator
  migrator -->|depends_on| postgresql_db
  ngrok -->|depends_on| cauldron

  kafka ---|volumes| kafka_data
  postgresql_db ---|volumes| postgres-data
```

Service sequence diagrams can be found [here][003]. Instead of running each
service independently, use:

```bash
docker compose -f docker-compose.infra.yml up    # or,
rake docker:compose:infra:up
```

Now you can access:

- Kafka UI: `http://127.0.0.1:8080/`
- Ngrok: `http://127.0.0.1:4040`
- PostgreSQL: `PGPASSWORD="${POSTGRES_PASSWORD}" psql -h localhost -p 5433 -U postgres -d devchain_webhook`

For PostgreSQL, `5433` is exposed in container to avoid conflicts with the
local PostgreSQL instance.

Logging for **kafka** and **kafka-ui** is set to `error` only. Due to development
purposes, both were producing too much information, little clean up required.

---

## Database Migrations

Add new migration:

```bash
migrate create -ext sql -dir migrations -seq <name>
```

[001]: https://brew.sh/
[002]: https://orbstack.dev/
[003]: https://github.com/devchain-network/cauldron/blob/main/SEQUENCE_DIAGRAMS.md
