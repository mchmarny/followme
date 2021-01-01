# followme

Monitor new daily Twitter followers/unfollowers across multiple accounts. Identify who doesn't follow you back and whom you don't follow back without handing over your entire Twitter account over to another on-line service.

## Why app vs service

While there are many services available to track your followers and unfollowers on-line, they all tend to require a full access to your Twitter account, including the ability to make changes to your profile, and tweet, like, or DM on your behalf. 

The followme app requires only read-only level access to your Twitter account, and uses only data already available publicly, and keeps your data stored locally, on your device.

## Install

The followme app is available on Mac, Linux, and Windows platforms. 

### Mac or Linux 

Install using Homebrew:

```shell
brew tap mchmarny/followme
brew install followme
```

Subsequent releases will be automatically picked up with `brew upgrade`.

### Windows 

> The choco package for followme is in works, stay tuned. For now install manually.

Download [latest release](https://github.com/mchmarny/followme/releases/latest) and place it in your path.

## Setup 

To use followme you will need Twitter API credentials (consumer key and secret):

1. Navigate to https://dev.twitter.com/apps/new and log in
2. Enter your app name (`followme` or similar) and description (leave URL empty)
3. Accept the TOS, and solve the CAPTCHA.
4. Submit the form by clicking the Create your Twitter Application
5. Copy the consumer key and consumer secret to use with followme app and worker

## Usage

> For both the follow me app and worker you can either provide the `--key` and `--secret` flags on each launch, or define the `TWITTER_CONSUMER_KEY` and `TWITTER_CONSUMER_SECRET` variables

### App

The followme app displays your Twitter follower data.

```shell
followme app --key $YOUR_TWITTER_CONSUMER_KEY \
             --secret $TWITTER_CONSUMER_SECRET
```

> List all flags supported by followme using `followme` without any arguments

The above command will launch followme app in your browser.

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