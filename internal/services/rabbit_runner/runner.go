package rabbitrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"rankode/internal/models"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	reqQueue  = "rankode-req"
	respQueue = "rankode-resp"
)

type TasksValidator interface {
	ValidateAndUpdate(ctx context.Context, data models.AttemptResponse)
}

type RabbitMQRunnerConfig struct {
	Login        string
	Password     string
	Host         string
	Port         int
	WorkersCount int
}

type RabbitMQRunner struct {
	cfg          RabbitMQRunnerConfig
	conn         *amqp.Connection
	producerChan *amqp.Channel
	consumerChan *amqp.Channel
	attemptsChan chan models.AttemptResponse
	closed       bool
	wg *sync.WaitGroup
	validator TasksValidator
}

func NewRabbitMQRunner(cfg RabbitMQRunnerConfig, validator TasksValidator) (*RabbitMQRunner, error) {
	return &RabbitMQRunner{cfg: cfg, wg: &sync.WaitGroup{}, validator: validator}, nil
}

func (r *RabbitMQRunner) Start() error {
	if err := r.connect(); err != nil {
		return err
	}
	if err := r.startProducer(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	if err := r.startConsumer(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	for i := 0; i < r.cfg.WorkersCount; i++ {
		r.wg.Add(1)
		go r.worker()
	}
	return nil
}

func (r *RabbitMQRunner) startProducer() error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}
	_, err = channel.QueueDeclare(reqQueue, false, false, false, false, nil)
	
	if err != nil {
		return err
	}
	r.producerChan = channel
	return nil
}

func (r *RabbitMQRunner) startConsumer() error {
	channel, err := r.conn.Channel()
	if err != nil {
		return err
	}
	queue, err := channel.QueueDeclare(respQueue, false, false, false, false, nil)
	if err != nil {
		return err
	}
	del, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	r.consumerChan = channel
	r.attemptsChan = make(chan models.AttemptResponse, 10)
	go r.listener(del)
	return nil
}

func (r *RabbitMQRunner) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", r.cfg.Login, r.cfg.Password, r.cfg.Host, r.cfg.Port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	errChan := make(chan *amqp.Error)
	conn.NotifyClose(errChan)
	r.conn = conn
	go func() {
		<-errChan
		r.Close()
		if r.closed {
			return
		}

		for {
			time.Sleep(time.Second * 15)
			err := r.Start()
			if err == nil {
				return
			}
		}
	}()
	return nil
}

func (r *RabbitMQRunner) listener(taskChan <-chan amqp.Delivery) {
	for data := range taskChan {
		var task models.AttemptResponse
		if err := json.Unmarshal(data.Body, &task); err != nil {
			slog.Error("invalid task message", "message", string(data.Body), "error", err)
			continue
		}
		r.attemptsChan <- task
	}
}

func (r *RabbitMQRunner) worker() {
	for task := range r.attemptsChan{
		r.validator.ValidateAndUpdate(context.Background(), task)
	}
}


func (r *RabbitMQRunner) Close() {
	close(r.attemptsChan)
	r.wg.Wait()
}

func (r *RabbitMQRunner) SendAttempt(data models.AttemptRequest) error {
	body, _ := json.Marshal(data)
	err := r.producerChan.Publish("", reqQueue, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})
	return err
}
