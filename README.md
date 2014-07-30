# <img src="https://i.cloudup.com/WSKggRp4ZX.svg" width=78 /> <br/> busl

busl - the bustle part of hustle.

a simple pubsub service that runs on Heroku.

[![Build Status](https://travis-ci.org/naaman/busl.svg?branch=master)](https://travis-ci.org/naaman/busl)

[![Deploy to Heroku](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

## usage

create a stream:

```
$ export STREAM_ID=$(curl busl.herokuapp.com/streams -X POST)
b7e586c8404b74e1805f5a9543bc516f
```

connect a consumer using the stream id:

```
$ curl busl.herokuapp.com/streams/$STREAM_ID
...
```

in a separate terminal, produce some data using the same stream id...

```
$ curl busl.herokuapp.com/streams/$STREAM_ID -X POST
```

...and you see the busl.
