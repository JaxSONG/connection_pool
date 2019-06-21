package pool

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type mqttConn struct {
	IsAlive   bool
	InitTime  time.Time
	Client    MQTT.Client
	IsTimeOut func() bool
}

func NewMqttConn(f func() bool, initTime time.Time) *mqttConn {
	return &mqttConn{
		IsTimeOut: f,
		InitTime:  initTime,
	}
}

func (mc *mqttConn) IsHealthy() bool {
	if !mc.IsTimeOut() {
		return false
	}
	mc.IsAlive = mc.Client.IsConnected()
	return mc.IsAlive
}

func (mc *mqttConn) Close() {
	mc.GetConn().(MQTT.Client).Disconnect(100)
}

func (mc *mqttConn) GetConn() interface{} {
	return mc.Client
}

func (mc *mqttConn) ReUse(ch chan ClientConn) {
	select {
	case ch <- mc:
	default:
		mc.Client.Disconnect(10)
	}
}
