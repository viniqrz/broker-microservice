package main

import (
	"encoding/json"
	"fmt"
	"sync"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/viniqrz/broker-microservice/internal/infra/kafka"
	"github.com/viniqrz/broker-microservice/internal/market/dto"
	"github.com/viniqrz/broker-microservice/internal/market/entity"
	"github.com/viniqrz/broker-microservice/internal/market/transformer"
)


func main() {
	ordersIn := make(chan *entity.Order)
	ordersOut := make(chan *entity.Order)
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	kafkaMsgChan := make(chan *ckafka.Message)
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": "host.docker.internal:9094",
		"group.id": "myGroup",
		"auto.offset.reset": "latest",
	}
	producer := kafka.NewProducer(configMap)
	consumer := kafka.NewConsumer(configMap, []string{"input"})

	go consumer.Consume(kafkaMsgChan)

	book := entity.NewBook(ordersIn, ordersOut, wg)
	go book.Trade()

	go func() {
		for kafkaMessage := range kafkaMsgChan {
			wg.Add(1)
			fmt.Println(string(kafkaMessage.Value))
			tradeInput := dto.TradeInput{}
			err := json.Unmarshal(kafkaMessage.Value, &tradeInput)
			if err != nil {
				panic(err)
			}
			orderFromKafka := transformer.TransformInput(tradeInput)
			ordersIn <- orderFromKafka
		}
	}()

	for orderFromChanOut := range ordersOut {
		orderOutput := transformer.TransformOutput(orderFromChanOut)
		orderOutputMessage, err := json.MarshalIndent(orderOutput, "", " ")
		if err != nil {
			fmt.Println(err)
		}
		producer.Publish(orderOutputMessage, []byte{007}, "output")
	}
}