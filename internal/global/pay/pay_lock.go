package pay

import "sync"

var withdrawOrderLockMap map[string]*sync.RWMutex = make(map[string]*sync.RWMutex)

// 上锁
func WithdrawOrderLock(orderId string) {
	lock, ok := withdrawOrderLockMap[orderId]
	if !ok {
		lock = &sync.RWMutex{}
		withdrawOrderLockMap[orderId] = lock
	}
	lock.TryLock()
}

// 解锁
func WithdrawOrderUnlock(orderId string, del bool) {
	lock, ok := withdrawOrderLockMap[orderId]
	if !ok {
		return
	}
	lock.Unlock()
	if del {
		delete(withdrawOrderLockMap, orderId)
	}
}
