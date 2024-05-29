# social-2-telego

## Comparing to my previous project

[social-2-telegram](https://github.com/Delnegend/social-2-telegram) but
- faster
- written in Go instead of Python
- run in headless docker, no need a browser
- interact through Telegram Bot, not terminal

## Set-up
- Install [Docker](https://docs.docker.com/get-docker/)
- Clone this repo
- Run `sh build.sh` to build the binary
- Modify the envs in `docker-compose.yml`
- `docker compose up -d` to start the bot

## Environment variables

| Name                        | Description                                                                                                | Example                                     | Default |
|-----------------------------|------------------------------------------------------------------------------------------------------------|---------------------------------------------|---------|
| `PORT`                      | the port the app listens to                                                                                | `8080`                                      | `8080`  |
|                             |                                                                                                            |                                             |         |
| `USE_WEBHOOK`               | whether to use webhook or long-polling                                                                     | `false`                                     |         |
| `WEBHOOK_DOMAIN`            | the domain to set the webhook, `/webhook` will be appended                                                 | `https://example.com`                       |         |
| `WEBHOOK_SECRET`            | the secret to verify incoming requests; only alphanumeric, underscore, and hyphen are allowed              | `_secret@123_`                              |         |
| `RETRY_SET_WEBHOOK_ATTEMPT` | # of attempts to set the webhook if it fails before giving up                                              | `5`                                         | `5`     |
|                             |                                                                                                            |                                             |         |
| `GET_UPDATES_INTERVAL`      | the interval to get updates in long-polling mode                                                           | `5s`                                        | `5s`    |
|                             |                                                                                                            |                                             |         |
| `BOT_TOKEN`                 | the Telegram Bot token                                                                                     | `123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11` |         |
| `ARTIST_DB_DOMAIN`          | the domain of the artist database, leave it empty to default to use the same social media as the given URL | `https://example.com/{username}`            |         |
| `ALLOWED_USERS`             | the list of allowed users, separated by `,`; leave it empty to allow all users                             | `user1,user2`                               |         |
| `TARGET_CHANNEL`            | the channel's handle to send the messages to (without `@`)                                                 | `myChannel`                                 |         |
|                             |                                                                                                            |                                             |         |
| `NUM_WORKERS`               | the number of concurrent workers to process the messages                                                   | `5`                                         | `5`     |

## Message format

```
https://x.com/foo/status/123, @bar Bar, #baz
```
- When splitting the message by `,`, there must be at least 1 element, at most 3 elements.
- First element must match the URL pattern
- Second/third element are the artist's name/username overwrite and hashtags. They are optional and the position can be exchanged.
- The artist's name/username overwrite element must start with `@` and the hashtags element must start with `#`.

## Update
```bash
docker down && git stash && git pull --rebase && git stash apply && docker up -d --build
```