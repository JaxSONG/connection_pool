package pool

import (
	"context"
	"fmt"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func Test_Pool(t *testing.T) {
	for i := 0; i < 100; i++ {
		foo(i)
	}
}

var p *Pool

func init() {
	p = NewPool(NewFactory(time.Minute), 3, 6, time.Minute)
}

func foo(idex int) {
	cli, err := p.Get(context.Background())
	if err != nil {
		panic(err)
	}
	c := cli.GetConn().(MQTT.Client)
	c.Publish("hello", 1, false, "hello"+fmt.Sprintf("%d", idex))
	defer cli.ReUse(p.clients)
}

func NewFactory(MaxLifeTime time.Duration) Factory {
	return func(times int) (ClientConn, error) {
		opts := MQTT.NewClientOptions().
			AddBroker("tcp://mqtt.develop.meetwhale.com:1883").
			SetCleanSession(true).
			SetOrderMatters(false).
			SetKeepAlive(60 * time.Second).
			SetAutoReconnect(true).
			SetClientID("mqtt-pool" + fmt.Sprintf("%d", times))
		fmt.Println("create times:", times)
		c := MQTT.NewClient(opts)
		if token := c.Connect(); token.Wait() && token.Error() != nil {
			return nil, token.Error()
		}
		createT := time.Now()
		var conn = NewMqttConn(func() bool {
			if time.Now().Sub(createT) > MaxLifeTime {
				return false
			}
			return true
		}, createT)
		conn.Client = c
		conn.IsAlive = true
		return conn, nil
	}

}
