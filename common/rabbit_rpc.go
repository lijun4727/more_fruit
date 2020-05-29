package common

import (
	context "context"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitRpcTimeoutError struct {
}

func (e *RabbitRpcTimeoutError) Error() string {
	return "Call rabbit rpc time out"
}

type RabbitParentCtxCancelError struct {
}

func (e *RabbitParentCtxCancelError) Error() string {
	return "Parent context cancel"
}

type RabbitAdrressMatchError struct {
}

func (e *RabbitAdrressMatchError) Error() string {
	return "don'n match previous address"
}

type RabbitRpc struct {
	amqpConn *amqp.Connection
	//amqpCh   *amqp.Channel
	mutex         sync.Mutex
	address       string
	context       context.Context
	contextCancle context.CancelFunc
}

func (rabRpc *RabbitRpc) Connect(address string) error {
	// if len(rabRpc.address) > 0 && rabRpc.address != address {
	// 	return &RabbitAdrressMatchError{}
	// }

	// if rabRpc.amqpConn != nil && !rabRpc.amqpConn.IsClosed() {
	// 	return nil
	// }

	// rabRpc.mutex.Lock()
	// defer rabRpc.mutex.Unlock()

	var err error
	rabRpc.amqpConn, err = amqp.Dial(address)
	if err != nil {
		return err
	}
	// rabRpc.amqpCh, err = rabRpc.amqpConn.Channel()
	// if err != nil {
	// 	rabRpc.amqpConn.Close()
	// }
	//err = sr.amqpCh.Confirm(false)
	rabRpc.address = address
	rabRpc.context, rabRpc.contextCancle = context.WithCancel(context.Background())
	return err
}

func (rabRpc *RabbitRpc) ResetConnect() error {
	if rabRpc.amqpConn != nil && !rabRpc.amqpConn.IsClosed() {
		return nil
	}
	rabRpc.mutex.Lock()
	defer rabRpc.mutex.Unlock()
	var err error
	rabRpc.amqpConn, err = amqp.Dial(rabRpc.address)
	return err
}

func (rabRpc *RabbitRpc) CallRpc(routeKey string, body string, timeout time.Duration) ([]byte, error) {
	amqpCh, err := rabRpc.amqpConn.Channel()
	if err != nil {
		return nil, err
	}
	defer amqpCh.Close()
	amqpQue, err := amqpCh.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}
	defer amqpCh.QueueDelete(amqpQue.Name, false, false, false)

	msgs, err := amqpCh.Consume(
		amqpQue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return nil, err
	}

	err = amqpCh.Publish(
		"",       // exchange
		routeKey, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: amqpQue.Name,
			DeliveryMode:  amqp.Persistent,
			ReplyTo:       amqpQue.Name,
			Body:          []byte(body),
		})
	if err != nil {
		return nil, err
	}

	var t time.Timer
	if timeout > 0 {
		t = *time.NewTimer(timeout)
	}
	select {
	case msg := <-msgs:
		msg.Ack(false)
		return msg.Body, nil
	case <-t.C:
		return nil, &RabbitRpcTimeoutError{}
	case <-rabRpc.context.Done():
		return nil, &RabbitParentCtxCancelError{}
	}
}

type CallbackFun func(context *context.Context, jsonParam string) string

func (rabRpc *RabbitRpc) ResponceRpc(routeKey string, callback CallbackFun) error {
	amqpCh, err := rabRpc.amqpConn.Channel()
	if err != nil {
		return err
	}
	defer amqpCh.Close()

	amqpQue, err := amqpCh.QueueDeclare(
		routeKey, // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	threadNum := runtime.NumCPU()
	err = amqpCh.Qos(
		threadNum, // prefetch count
		0,         // prefetch size
		false,     // global
	)
	if err != nil {
		return err
	}
	msgs, err := amqpCh.Consume(
		amqpQue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(threadNum)
	for i := 0; i < threadNum; i++ {
		go func() {
			for {
				select {
				case msg := <-msgs:
					responceCode := callback(&rabRpc.context, string(msg.Body))
					err = amqpCh.Publish(
						"",          // exchange
						msg.ReplyTo, // routing key
						false,       // mandatory
						false,       // immediate
						amqp.Publishing{
							ContentType:   "text/plain",
							CorrelationId: msg.CorrelationId,
							Body:          []byte(responceCode),
						})
					if err != nil {
						log.Printf("amqpCh.Publish fail:%v", err)
					}
					msg.Ack(false)
				case <-rabRpc.context.Done():
					wg.Done()
					return
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (rabRpc *RabbitRpc) Close() {
	if rabRpc.contextCancle != nil {
		rabRpc.contextCancle()
	}
	if rabRpc.amqpConn != nil {
		rabRpc.amqpConn.Close()
	}
}
