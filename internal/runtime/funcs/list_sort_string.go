package funcs

import (
	"context"

	"github.com/nevalang/neva/internal/runtime"
	"golang.org/x/exp/slices"
)

type listSortString struct{}

func (p listSortString) Create(io runtime.IO, _ runtime.Msg) (func(ctx context.Context), error) {
	dataIn, err := io.In.Single("data")
	if err != nil {
		return nil, err
	}

	resOut, err := io.Out.Single("res")
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context) {
		for {
			data, ok := dataIn.Receive(ctx)
			if !ok {
				return
			}

			clone := slices.Clone(data.List())
			slices.SortFunc(clone, func(i, j runtime.Msg) bool {
				return i.Str() < j.Str()
			})

			if !resOut.Send(ctx, runtime.NewListMsg(clone)) {
				return
			}
		}
	}, nil
}
