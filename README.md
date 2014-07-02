#busl

busl - the bustle part of hustle.

a simple pubsub service that runs on Heroku.

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
