package mq

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/meshplus/bitxhub-kit/log"
	"github.com/streadway/amqp"
)

var logger = log.NewWithModule("mq")

type MessageHandler interface {
	HandleMessage([]byte)
}

type Consumer struct {
	uri          string
	queueName    string
	exchange     string
	exchangeType string
	routingKey   string
	logger       hclog.Logger
	conn         *amqp.Connection
	channel      *amqp.Channel
	tag          string
	close        chan struct{}
	msgH         MessageHandler
}

func NewConsumer(opts ...Option) (*Consumer, error) {
	config, err := generateConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("wrong config: %w", err)
	}

	c := &Consumer{
		uri:          config.uri,
		queueName:    config.queueName,
		exchange:     config.exchange,
		logger:       config.logger,
		exchangeType: config.exchangeType,
		routingKey:   config.routingKey,
		conn:         nil,
		channel:      nil,
		tag:          "simple-consumer",
		close:        make(chan struct{}),
		msgH:         config.handler,
	}

	return c, nil
}

func (c *Consumer) Start() error {
	go func() {
		for {
			time.Sleep(3 * time.Second)
			if err := c.setup(); err != nil {
				logger.Infof("Set up MQ consumer: %s", err.Error())
				continue
			}

			c.logger.Info("MQ consumer started")
			<-c.close
		}
	}()

	return nil
}

func (c *Consumer) setup() error {
	var err error

	c.conn, err = amqp.Dial(c.uri)
	if err != nil {
		c.logger.Error("amqp dial: %w", err)
		return fmt.Errorf("amqp dial: %w", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.logger.Error("create channel error: %w", err)
		return fmt.Errorf("create channel error: %w", err)
	}

	if err = c.channel.ExchangeDeclare(c.exchange, c.exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	queue, err := c.channel.QueueDeclare(c.queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}
	c.logger.Info("MQ queue started:", queue.Name)

	if err = c.channel.QueueBind(queue.Name, c.routingKey, c.exchange, false, nil); err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	deliveries, err := c.channel.Consume(queue.Name, c.tag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue consume: %w", err)
	}
	c.logger.Info("MQ queue deliveries:", len(deliveries))

	go c.handle(deliveries)

	return nil
}

func (c *Consumer) Stop() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer cancel: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("amqp connection close: %w", err)
	}

	return nil
}

func (c *Consumer) handle(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		c.msgH.HandleMessage(d.Body)
		err := d.Ack(false)
		if err != nil {
			c.logger.Error("delivery ack: %s", err)
		}
	}

	c.logger.Info("deliveries channel closed")

	c.close <- struct{}{}
}
