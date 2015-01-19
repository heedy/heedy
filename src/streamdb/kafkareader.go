package main

import (
    "github.com/Shopify/sarama"
    "fmt"
    "os"
    "bufio"
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
                panic(0)
            }
            defer client.Close()

            topics,err := client.Topics()
            if (err!= nil) {
                fmt.Printf("Error: %s\n",err)
                panic(0)
            }
            fmt.Printf("Topics: %s\n",topics)
            for t := range topics {
                partitions,err := client.Partitions(topics[t])
                if (err != nil) {
                    fmt.Printf("Error: %s\n",err)
                    } else {
                        for p := range partitions {
                            go TopicReader(topics[t],partitions[p],client)
                        }
                    }
                }

                reader := bufio.NewReader(os.Stdin)
                fmt.Print("Press Enter to stop")
                _, _ = reader.ReadString('\n')

            }
            
