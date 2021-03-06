package system

import (
	"sync"
	"time"
)

type SmartLocker interface {
	//timeout lock
	TryLock(timeout time.Duration) bool
}

type PoppyLock struct {
	sync.Mutex
}

func (fl *PoppyLock) TryLock(timeout time.Duration) bool {
	//Open a coroutine to lock, if the lock is successful, notify the current coroutine
	//The current coroutine starts the timer, if the time is up, it will automatically exit
	getLock := make(chan struct{}, 1)
	timeoutLock := make(chan struct{}, 1)
	time.AfterFunc(timeout, func() {
		timeoutLock <- struct{}{}
	})
	go func() {
		fl.Lock()
		getLock <- struct{}{}
		//registered callback hook
		//why i add this ?
		//because this apply locker's routine will get lock.
		//If I do not release the lock, the lock will be occupied
		//Being able to execute here means that the lock must be acquired
		select {
		//The lock is acquired in the case of timeout and must be released
		case <-timeoutLock:
			fl.Unlock()
		default:
			//FIXED 2-check Since select will give priority to other channels, if not, select default. Therefore, double check is not required
			//if promise {
			//	fl.Unlock()
			//	return
			//}
			return
		}
	}()
	select {
	case <-getLock:
		//if get lock, the caller know should release.
		return true
	case <-timeoutLock:
		//add empty to notify that apply locker routine should release the lock
		timeoutLock <- struct{}{}
		return false
	}
}
