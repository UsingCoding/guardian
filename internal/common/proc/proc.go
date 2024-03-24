package proc

type Proc interface {
	Start() error
	Stop() error
}

func NewProc(
	start func() error,
	stop func() error,
) Proc {
	return &proc{
		start: start,
		stop:  stop,
	}
}

type proc struct {
	start func() error
	stop  func() error
}

func (p *proc) Start() error {
	return p.start()
}

func (p *proc) Stop() error {
	return p.stop()
}
