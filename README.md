![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)
![Go](https://img.shields.io/github/go-mod/go-version/devchain-network/cauldron)
[![codecov](https://codecov.io/github/devchain-network/cauldron/graph/badge.svg?token=LAUHZBW12F)](https://codecov.io/github/devchain-network/cauldron)


# cauldron

A dynamic hub where incoming webhooks brew, process, and transform seamlessly.

---

## Requirements for Local Development

- `go` - current version: `1.23.4`
- `ruby` - optional, if you add/implement `rake` tasks, will lint your ruby code.
- `direnv` - manages your environment variables.
- `pre-commit` - use `pre-commit install` if you plan to contribute!
- `docker` - you can run external services with `docker-compose`
- `ngrok` or equivalent tool if you want to receive webhooks locally.

Optional; you can install `robocop` for linting `Rakefile`:

```bash
bundle config set --local path 'ruby-vendor/bundle' --local bin 'bin'
bundle
```

### Environment Variables

Environment variables for server:

| Variable | Description | Default |
|:---------|:------------|---------|
| `LISTEN_ADDR` | Server listen address | `":8000"` |
| `LOG_LEVEL` | Logging level, Valid values are: `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"` | `"INFO"` |
| `GITHUB_HMAC_SECRET` | HMAC secret value for GitHub’s webhooks. | `""` |
| `KCP_TOPIC_GITHUB` | Kafka consumer/producer topic name for GitHub Webhooks | `""` |
| `KCP_BROKERS` | Kafka consumer/producer brokers list, comma separated | `"127.0.0.1:9094"` |
| `KP_PRODUCER_QUEUE_SIZE` | Size of default Kafka message producer queue size | `100` |

Environment variables GitHub consumer:

| Variable | Description | Default |
|:---------|:------------|---------|
| `LOG_LEVEL` | Logging level, Valid values are: `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"` | `"INFO"` |
| `KC_PARTITION` | Consumer partition number | `0` |
| `KC_TOPIC` | Topic to subscribe | `""` |
| `KCP_BROKERS` | Kafka consumer/producer brokers list, comma separated | `"127.0.0.1:9094"` |
| `KC_DIAL_TIMEOUT` | Initial connection timeout used by broker | "`30s`" (seconds) |
| `KC_READ_TIMEOUT` | Response timeout used by broker | "`30s`" (seconds) |
| `KC_WRITE_TIMEOUT` | Transmit timeout used by broker | "`30s`" (seconds) |
| `KC_BACKOFF` | Backoff value for retries | "`2s`" (seconds) |
| `KC_MAX_RETRIES` | Maximum retry | `10` |

Example `.envrc`:

```bash
# for ruby/rake only
export PATH="bin:${PATH}"

export LISTEN_ADDR=":8000"
export LOG_LEVEL="INFO"
export GITHUB_HMAC_SECRET="<secret>"

# kafka consumer/producer shared values.
export KCP_TOPIC_GITHUB="github"
export KCP_BROKERS="127.0.0.1:9094"

# kafka producer values.
export KP_PRODUCER_QUEUE_SIZE=100

# kafka github consumer values.
export KC_PARTITION="0"
export KC_TOPIC="${KP_TOPIC_GITHUB}"
export KC_DIAL_TIMEOUT="30s"
export KC_READ_TIMEOUT="30s"
export KC_WRITE_TIMEOUT="30s"
export KC_BACKOFF="2s"
export KC_MAX_RETRIES="10"
```

---

## Installation

```bash
cd /path/to/development
git clone git@github.com:devchain-network/cauldron.git
cd cauldron/
go mod download
```

Run Kafka from docker with docker-compose:

```bash
docker compose -f docker-compose.local.yml up

# kafka-ui is available at: http://localhost:8080/
```

Run the server:

```bash
go run cmd/server/main.go
```

Run the GitHub webhook Kafka consumer:

```bash
go run cmd/githubconsumer/main.go
```

---

## Usage

@wip

---

## Docker

@wip

---

## Rake Tasks

```bash
rake -T

rake default                    # default task, runs server
rake run:compose:down           # run docker compose down
rake run:compose:up             # run docker compose up
rake run:kafka:github:consumer  # run kafka github consumer
rake run:server                 # run server
```

---

## License

This project is licensed under MIT

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/devchain-network/cauldron/blob/main/CODE_OF_CONDUCT.md
