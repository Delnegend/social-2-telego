# social-2-telego

## Comparing to my previous project

<table>
    <tr>
        <th />
        <th><a href="https://github.com/Delnegend/social-2-telegram">social-2-telegram</a></th>
        <th>social-2-telego</th>
    </tr>
    <tr>
        <td>How it works</td>
        <td>0. spins up a browser<br>1. takes a post URL<br>2. scrapes the content<br>3. composes a message<br>4. sends to Telegram</td>
        <td>0. spin up a Telegram bot (can run 24/7)<br>1. takes a post URL with some optional additional content (artist's name/username, hashtags)<br>2. pretty formatting<br>3. sends to Telegram</td>
    </tr>
    <tr>
        <td>Ease of use</td>
        <td>- requires a computer<br>- 20-25 seconds delay between each post URL due to scraping, social media links re-validation, asking for additional information</td>
        <td>Just send the link to the Telegram bot</td>
    </tr>
    <tr>
        <td>Advantages</td>
        <td>- Dependency-free, everything included in the Telegram post, no need for something to run 24/7</td>
        <td>- Fast, much more convenient, easier to update information for artists if</td>
    </tr>
    <tr>
        <td>Disadvantages</td>
        <td>- Slow, requires a Windows machine, a little bit clunky to set up</td>
        <td>- Relies on external services for viewing (<a href="https://github.com/FixTweet/FxTwitter">fxtwitter</a> and this bot for the preview)<br>- Require to maintaining a separated <a href="https://linktr.ee">linktr.ee</a>-like website</td>
    </tr>

</table>

## Set-up
- Install [Docker](https://docs.docker.com/get-docker/)
- Clone this repo
- Run `sh build.sh` to build the binary
- Rename `.env.example` to `.env` and fill in the required information
- `docker compose up -d` to start the bot

## Message format

```
https://x.com/foo/status/123, @bar Bar, #baz
```
- When splitting the message by `,`, there must be at least 1 element, at most 3 elements.
- First element must match the URL pattern
- Second/third element are the artist's name/username overwrite and hashtags. They are optional and the position can be exchanged.
- The artist's name/username overwrite element must start with `@` and the hashtags element must start with `#`.