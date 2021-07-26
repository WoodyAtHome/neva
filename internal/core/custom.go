package core

type customModule struct {
	deps    Deps
	in      InportsInterface
	out     OutportsInterface
	workers Workers
	net     Net
}

func (cm customModule) Interface() Interface {
	return Interface{
		In:  cm.in,
		Out: cm.out,
	}
}

type Workers map[string]string

type Net []Subscription

type Subscription struct {
	Sender    PortPoint
	Recievers []PortPoint
}

type PortPoint struct {
	Node string
	Port string
}

func NewCustomModule(
	deps Deps,
	in InportsInterface,
	out OutportsInterface,
	workers Workers,
	net Net,
) Module {
	return customModule{
		deps:    deps,
		in:      in,
		out:     out,
		workers: workers,
		net:     net,
	}
}
