package compiler

import (
	"fmt"

	"github.com/emil14/stream/internal/compiler/program"
	rprog "github.com/emil14/stream/internal/runtime/program"
)

type compiler struct {
	parser     Parser
	validator  Validator
	translator Translator
	coder      Coder
}

func (c compiler) Compile(src []byte) ([]byte, error) {
	mod, err := c.parser.Parse(src)
	if err != nil {
		return nil, err
	}

	if err := c.validator.Validate(mod); err != nil {
		return nil, err
	}

	prog := program.Program{
		Components: map[string]program.Component{
			"root": mod,
		},
	}

	bb, err := c.coder.Code(c.translator.Translate(prog))
	if err != nil {
		return nil, err
	}

	return bb, nil
}

type Coder interface {
	Code(rprog.Program) ([]byte, error)
}

func New(p Parser, v Validator, t Translator, c Coder) (compiler, error) {
	if p == nil || v == nil || t == nil || c == nil {
		return compiler{}, fmt.Errorf("failed to build compiler")
	}

	return compiler{
		parser:     p,
		validator:  v,
		translator: t,
		coder:      c,
	}, nil
}

func MustNew(p Parser, v Validator, t Translator, c Coder) compiler {
	cmp, err := New(p, v, t, c)
	if err != nil {
		panic(err)
	}
	return cmp
}
