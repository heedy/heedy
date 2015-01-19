package main

import (
    "github.com/Shopify/sarama"
    "fmt"
)


func TopicReader(topic string, partition int32, client *sarama.Client) {
    fmt.Printf("RunTopic: %s %d\n",topic,partition)
    cfg := sarama.NewConsumerConfig()
    cons,err := sarama.NewConsumer(client,topic,partition,"writer",cfg)
    if err != nil {
        fmt.Printf("ErrorT: %s\n",err)
    } else {
        defer cons.Close()
        for e := range cons.Events() {
            fmt.Printf("MSG: %s.%d M=%s\n",e.Topic,e.Partition,string(e.Value))
        }

    }
}


func main() {
    cfg := sarama.NewClientConfig()
    client,err := sarama.NewClient("writer",[]string{"localhost:1336"},cfg)
    if err != nil {
        fmt.Printf("Error: %s\n",err)
        panic(err)
    }
    defer client.Close()

    pcfg := sarama.NewProducerConfig()
    producer, err := sarama.NewProducer(client, pcfg)
    if err != nil {
        panic(err)
    }
    defer producer.Close()

    producer.Input() <-&sarama.MessageToSend{Topic: "i", Key: nil, Value: sarama.StringEncoder("testing 1234")}
}
