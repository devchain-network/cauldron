![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)
![Go](https://img.shields.io/github/go-mod/go-version/devchain-network/cauldron)
[![codecov](https://codecov.io/github/devchain-network/cauldron/graph/badge.svg?token=LAUHZBW12F)](https://codecov.io/github/devchain-network/cauldron)
[![Build and push Cauldron](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron.yml/badge.svg)](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron.yml)
[![Build and push Cauldron Migrator](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron-migrator.yml/badge.svg)](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron-migrator.yml)
[![Build and push Cauldron GitHub Consumer](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron-github-comsumer.yml/badge.svg)](https://github.com/devchain-network/cauldron/actions/workflows/build-push-cauldron-github-comsumer.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/devchain-network/cauldron)](https://goreportcard.com/report/github.com/devchain-network/cauldron)


# cauldron

A dynamic hub where incoming webhooks brew, process, and transform seamlessly.

More information will be provided soon.

---

## Supported GitHub Webhook Events

https://docs.github.com/en/webhooks/webhook-events-and-payloads

Applies to repositories and organizations.

| Event | Description |
|:------|:-----|
| `commit_comment` | Commit or diff commented on |
| `create` | Branch or Tag creation |
| `delete` | Branch or Tag deletion |
| `fork` | Repository forked |
| `gollum` | Wiki page updated |
| `issue_comment` | Issue comment created, edited, or deleted |
| `issues` | Issue opened, edited, deleted, transferred, pinned, unpinned, closed, reopened, assigned, unassigned, labeled, unlabeled, milestoned, demilestoned, locked, or unlocked |
| `pull_request` | Pull request assigned, auto merge disabled, auto merge enabled, closed, converted to draft, demilestoned, dequeued, edited, enqueued, labeled, locked, milestoned, opened, ready for review, reopened, review request removed, review requested, synchronized, unassigned, unlabeled, or unlocked |
| `pull_request_review_comment` | Pull request diff comment created, edited, or deleted |
| `pull_request_review` | Pull request review submitted, edited, or dismissed |
| `push` | Git push to a repository |
| `release` | Release created, edited, published, unpublished, or deleted |
| `star` | A star is created or deleted from a repository |
| `watch` | User stars a repository |

---

## License

This project is licensed under MIT

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/devchain-network/cauldron/blob/main/CODE_OF_CONDUCT.md
