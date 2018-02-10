package pubsub

import "github.com/garyburd/redigo/redis"

type Publisher struct {
  pool *redis.Pool
  message chan *redisMsg
}

func NewPublisher(pool *redis.Pool) *Publisher {
  return &Publisher{
    pool: pool,
    message: make(chan *redisMsg, 10),
  }
}

func (ps *Publisher) Run() error {
  conn := ps.pool.Get()
  defer conn.Close()

  for data := range ps.message {
    if err := conn.Send("PUBLISH", data.channel, data.message); err != nil {
      return err
    }
    if err := conn.Flush(); err != nil {
      return err
    }
  }
  return nil
}

func (ps *Publisher) Stop() {
  close(ps.message)
}

func (ps *Publisher) Publish(channel string, message []byte) {
  ps.message <- &redisMsg{channel: channel, message: message}
}
