# social-2-telego

## Comparing to my previous project

[social-2-telegram](https://github.com/Delnegend/social-2-telegram) but
- faster
- written in Go instead of Python
- run in headless docker, no need a browser
- interact through Telegram Bot, not terminal

## Prerequisites
- Docker
- A Telegram bot with its token
- ArtistDB

## Installation
- Clone this repo
- Rename `docker-compose.example.yml` to `docker-compose.yml`
- Modify the environment variables in `docker-compose.yml`
- `docker compose up -d` to start the bot

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