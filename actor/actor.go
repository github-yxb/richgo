package actor

type Actor struct {
	taskQueue chan func()
	cancelCh chan struct{}
}

func (a *Actor) Call(t func() interface{}) interface{} {
	retCh := make(chan interface{}, 1)
	a.taskQueue <- func() {
		retCh <- t()
	}
	return <-retCh
}

func (a *Actor) Post(t func()) {
	a.taskQueue <- t
}

func (a *Actor) Stop() {
	select {
		case a.cancelCh <- struct{}{}:
		default:
			return
		}
}


func (a *Actor) Start() {
	defer func() {
		//if err := recover(); err != nil {
		//
		//}
		close(a.taskQueue)
		close(a.cancelCh)
	}()

	a.taskQueue = make(chan func())
	a.cancelCh = make(chan struct{}, 1)

	for {
		select {
			case t := <- a.taskQueue:
				t()
			case <- a.cancelCh:
				return
		}
	}
}