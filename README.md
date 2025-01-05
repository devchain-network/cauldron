![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)
![Go](https://img.shields.io/github/go-mod/go-version/devchain-network/cauldron)
[![codecov](https://codecov.io/github/devchain-network/cauldron/graph/badge.svg?token=LAUHZBW12F)](https://codecov.io/github/devchain-network/cauldron)


# cauldron

A dynamic hub where incoming webhooks brew, process, and transform seamlessly.

---

## Requirements

@wip

---

## Requirements for Local Development

- `go`
- `ruby`
- `direnv`
- `pre-commit`

Example `.envrc`

```bash
export PATH="bin:${PATH}"
```

```bash
bundle config set --local path 'ruby-vendor/bundle' --local bin 'bin'
bundle
```

@wip


### Environment Variables

| Variable | Description | Default |
|:---------|:------------|---------|
| `LISTEN_ADDR` | Server listen address | `":8000"` |
| `LOG_LEVEL` | Logging level, Valid values are: `"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"` | `"INFO"` |
| `GITHUB_HMAC_SECRET` | HMAC secret value for GitHub’s webhooks. | `""` |

---

## Installation

@wip

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

rake default             # default task, runs server
rake run:compose:down    # run docker compose down
rake run:compose:up      # run docker compose up
rake run:kafka:consumer  # run kafka consumer
rake run:server          # run server
```

---

## License

This project is licensed under MIT

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/devchain-network/cauldron/blob/main/CODE_OF_CONDUCT.md
