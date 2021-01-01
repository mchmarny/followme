# followme

Monitor new Twitter followers/unfollowers across multiple accounts and identify who doesn't follow you back and whom you don't follow back without handing over your entire Twitter account over to another service. 

## Why followme app

While there are a few services available on-line to track your followers and unfollowers, they all tend to require a full write access to your Twitter account, including make changes to your profile and tweet, like, or DM on your behalf. 

followme app requires only read-only access to your Twitter account, uses only public data already available about your account, and keeps your data locally, on your device. 

## Install

### Mac or Linux 

Install followme with Homebrew:

```shell
brew tap mchmarny/followme
brew install followme
```

Subsequent releases will be automatically picked up with `brew upgrade`.

### Windows 

Download [latest release](https://github.com/mchmarny/followme/releases/latest) and place it in your path.

## Setup 

To use followme you will need Twitter consumer key (API key) and consumer secret:

1. Navigate to https://dev.twitter.com/apps/new and log in
2. Enter your `followme` or similar and description (leave URL empty)
3. Accept the TOS, and solve the CAPTCHA.
4. Submit the form by clicking the Create your Twitter Application
5. Copy the consumer key and consumer secret to use with followme app and worker

## Usage

> For both app and the worker you can either provide the `--key` and `--secret` flags of define the `TWITTER_CONSUMER_KEY` and `TWITTER_CONSUMER_SECRET` variables

### App

The followme app displays your Twitter follower data.

```shell
followme app --key $YOUR_TWITTER_CONSUMER_KEY \
             --secret $TWITTER_CONSUMER_SECRET \
             --port 8080
```

### Worker 

The followme worker updates your Twitter follower data. You can run it 1-2 times a day using cron.

```shell
followme worker --key $YOUR_TWITTER_CONSUMER_KEY \
                --secret $TWITTER_CONSUMER_SECRET
```


## Disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.

## License

This software is released under the [Apache v2](./LICENSE)