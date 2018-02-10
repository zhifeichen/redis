package pubsub

import (
  "context"
  "github.com/garyburd/redigo/redis"
)

type redisMsg struct {
  channel string
  message []byte
}

type Subscriber struct {
  pool *redis.Pool
  channels []string
  message chan *redisMsg
}

func NewSubscriber(pool *redis.Pool, channels ...string) *Subscriber  {
  return &Subscriber{
    pool:    pool,
    channels: channels,
    message: make(chan *redisMsg, 10),
  }
}

func (ss *Subscriber) Run(ctx context.Context, onMessage func(channel string, data []byte) error) error {
  conn := ss.pool.Get()
  defer conn.Close()

  psc := redis.PubSubConn{Conn: conn}
  if err := psc.Subscribe(redis.Args{}.AddFlat(ss.channels)...); err != nil {
    return err
  }

  done := make(chan error, 1)

  // Start a goroutine to receive notifications from the server.
  go func() {
    for {
      switch n := psc.Receive().(type) {
      case error:
        done <- n
        return
      case redis.Message:
        ss.message <- &redisMsg{channel: n.Channel, message: n.Data}
      case redis.Subscription:
        switch n.Count {
        case 0:
          // Return from the goroutine when all channels are unsubscribed.
          done <- nil
          return
        }
      }
    }
  }()
  var err error
loop:
  for err == nil {
    select {
    case msg := <-ss.message:
      onMessage(msg.channel, msg.message)
    case <-ctx.Done():
      break loop
    case err := <-done:
      // Return error from the receive goroutine.
      return err
    }
  }

  // Signal the receiving goroutine to exit by unsubscribing from all channels.
  psc.Unsubscribe()

  // Wait for goroutine to complete.
  return <-done
}
