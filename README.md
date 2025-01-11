# Reddit Discord Verification Bot

This is pretty rough around the edges and probably needs some touching up, but you should get the gist.

### How to Run

1. Make sure you have `make` installed in whatever CLI you're using
2. Make sure you have [Docker](https://docs.docker.com/engine/install/) and [Docker-Compose](https://docs.docker.com/compose/install/) installed
3. Fill out the .env config file
    1. DISCORD_BOT_TOKEN should be the token generated for whatever discord account you make from their [application portal](https://discord.com/developers/applications)
    2. REDDIT_USERNAME / REDDIT_PASSWORD should be the username and password for the Reddit account bot
    3. REDDIT_CLIENT_ID and REDDIT_SECRET_ID should be the client id and secret for an app *created on the bot account* in [Reddit's 3rd party app portal](https://www.reddit.com/prefs/apps/)
    4. MESSAGE_SUBJECT is just the message subject for the message users will be asked to send to the bot
4. Run: `make start`
5. Invite the bot to whatever server you want using Discord's oauth2 url generator
    1. You can find the docs [here](https://discord.com/developers/docs/topics/oauth2#authorization-code-grant)
    2. You can find it in Discord's application portal, link on line 3.1
    3. If you care about security, grant as small of a permission scope as possible. If you don't, give it `bot.administrator` permissions

I'll update this repo as people report issues, but in its current state, it's totally un-tested. Good luck.
