# hayaoki_bot

This is a slack bot for club-hayaoki.

It manages the hayaoki activity of the user in club-hayaoki.

## Deploy project to GCP

Install gcloud command from [this page](https://cloud.google.com/sdk/downloads).


Install App Engine Component.


```
$ sudo gcloud components install app-engine-go  
```

Set these slack parameters to Cloud Datastore.

| kind | key name string | data |
| --- | --- | --- |
| slack  | slash_token | {value : "[your slash command token]"} |
| slack  | bot_token | {value : "[your bot token]"} |

Deploy application.

```
$ git clone git@github.com:tdaira/hayaoki_bot.git
$ cd hayaoki_bot/app
$ gcloud app deploy
```
