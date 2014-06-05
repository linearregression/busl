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

## stupid ruby example

consuming:

```ruby
require 'excon'
 
streamer = lambda do |chunk, remaining_bytes, total_bytes|
    puts chunk
end
 
res = Excon.get(
  "http://busl.herokuapp.com/streams/#{ARGV[0]}", 
  response_block: streamer
)
```

producing:

```ruby
require 'excon'
 
file = File.open(ARGV[1], 'rb')
 
streamer = lambda do 
  file.read(Excon.defaults[:chunk_size]).to_s
end
 
res = Excon.post(
  "http://busl.herokuapp.com/streams/#{ARGV[0]}", 
  headers: { CONTENT_TYPE: 'application/octet-stream' },
  request_block: streamer
)
 
puts res.inspect
```

who uses ruby anyway
