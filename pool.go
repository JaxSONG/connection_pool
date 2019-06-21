// Package grpcpool provides a pool of grpc clients
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrClosed is the error when the client pool is closed
	ErrClosed = errors.New("grpc pool: client pool is closed")
	// ErrTimeout is the error when the client pool timed out
	ErrTimeout = errors.New("grpc pool: client pool timed out")
	// ErrAlreadyClosed is the error when the client conn was already closed
	ErrAlreadyClosed = errors.New("grpc pool: the connection was already closed")
	// ErrFullPool is the error when the pool is already full
	ErrFullPool = errors.New("grpc pool: closing a ClientConn into a full pool")
)

// Factory is a function type creating a grpc client
type Factory func(int) (ClientConn, error)

// Pool is the grpc client pool
type Pool struct {
	clients         chan ClientConn
	factory         Factory
	idleTimeout     time.Duration
	maxLifeDuration time.Duration
	mu              sync.RWMutex
	Id              int
}

func (pool *Pool) Close() {
	close(pool.clients)
	for c := range pool.clients {
		if c != nil {
			c.Close()
		}
	}
}
func (pool *Pool) Get(ctx context.Context) (ClientConn, error) {
	var wrapper ClientConn
	var err error
	select {
	case <-ctx.Done():
		return nil, ErrTimeout
	case wrapper = <-pool.clients:
	}
	fmt.Println("wrapper:", wrapper)
	if wrapper == nil || !wrapper.IsHealthy() {
		wrapper, err = pool.factory(pool.getId())
	}
	return wrapper, err
}
func (pool *Pool) getId() int {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.Id++
	return pool.Id
}
func NewPool(factory Factory, init int, cap int, maxLifeTime time.Duration) *Pool {
	p := &Pool{
		clients:         make(chan ClientConn, cap),
		maxLifeDuration: maxLifeTime,
		factory:         factory,
		Id:              0,
	}
	for i := 0; i < init; i++ {
		c, _ := factory(p.getId())
		p.clients <- c
	}
	for i := 0; i < cap-init; i++ {
		var c ClientConn
		p.clients <- c
	}
	return p
}

// ClientConn is the wrapper for a grpc client conn
type ClientConn interface {
	IsHealthy() bool
	ReUse(chan ClientConn)
	Close()
	GetConn() interface{}
}
