package distributed

import (
	"time"

	"github.com/nexuer/corex/distributed/registry"
)

type Locker interface {
	Lock(key string, duration ...time.Duration) error
	TryLock(key string, duration ...time.Duration) (bool, error)
	Unlock(key string) error
}

type Distributed interface {
	registry.Registrar
	registry.Discovery

	Locker
}
