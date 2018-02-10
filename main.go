package main

import (
  "bufio"
  "context"
  "fmt"
  _ "github.com/joho/godotenv/autoload"
  "github.com/zhifeichen/redis/pubsub"
  "os"
)

var channel  = "alarm:fire"

func onMessage(channel string, message []byte) error {
  fmt.Printf("%s <= %s\n", channel, string(message))
  return nil
}

func main() {
  addr := os.Getenv("REDIS_ADDR")
  pubsub.InitPool(addr)
  defer pubsub.ClosePool()

  subscriber := pubsub.NewSubscriber(pubsub.Pool(), channel)
  publisher := pubsub.NewPublisher(pubsub.Pool())
  ctx, cancel := context.WithCancel(context.Background())

  go func() {
    if err := subscriber.Run(ctx, onMessage); err != nil {
      fmt.Println(err.Error())
      return
    }
  }()

  go func() {
    if err := publisher.Run(); err != nil {
      return
    }
  }()
  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
    input := scanner.Text()
    if input == "byte" {
      break
    }
    publisher.Publish(channel, []byte(input))
  }
  cancel()
  publisher.Stop()
}
