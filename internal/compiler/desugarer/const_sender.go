package desugarer

import (
	"fmt"
	"sync/atomic"

	"github.com/nevalang/neva/internal/compiler"
	src "github.com/nevalang/neva/internal/compiler/sourcecode"
	"github.com/nevalang/neva/internal/compiler/sourcecode/core"
	ts "github.com/nevalang/neva/internal/compiler/sourcecode/typesystem"
)

var emitterComponentRef = core.EntityRef{
	Pkg:  "builtin",
	Name: "New",
}

type handleLiteralSenderResult struct {
	constName                  string
	handleConstRefSenderResult // conceptually incorrrect but convenient to reuse
}

type handleConstRefSenderResult struct {
	connectionWithoutConstSender src.Connection
	emitterNodeName              string
	emitterNode                  src.Node
}

// In the future compiler can operate in concurrently
var (
	virtualEmittersCount atomic.Uint64
	virtualConstCount    atomic.Uint64
)

func (d Desugarer) handleLiteralSender(
	conn src.Connection,
) (
	handleLiteralSenderResult,
	*compiler.Error,
) {
	constCounter := virtualConstCount.Load()
	virtualConstCount.Store(constCounter + 1)
	constName := fmt.Sprintf("virtual_const_%d", constCounter)

	// we can't call d.handleConstRefSender()
	// because our virtual const isn't in the scope

	emitterNode := src.Node{
		Directives: map[src.Directive][]string{
			compiler.BindDirective: {constName},
		},
		EntityRef: emitterComponentRef,
		TypeArgs: []ts.Expr{
			conn.
				Normal.SenderSide.
				Const.
				Message.
				TypeExpr,
		},
	}

	emitterCounter := virtualEmittersCount.Load()
	virtualEmittersCount.Store(emitterCounter + 1)
	emitterNodeName := fmt.Sprintf("virtual_emitter_%d", emitterCounter)

	emitterNodeOutportAddr := src.PortAddr{
		Node: emitterNodeName,
		Port: "msg",
	}

	return handleLiteralSenderResult{
		constName: constName,
		handleConstRefSenderResult: handleConstRefSenderResult{
			connectionWithoutConstSender: src.Connection{
				Normal: &src.NormalConnection{
					SenderSide: src.ConnectionSenderSide{
						PortAddr:  &emitterNodeOutportAddr,
						Selectors: conn.Normal.SenderSide.Selectors,
						Meta:      conn.Normal.SenderSide.Meta,
					},
					ReceiverSide: conn.Normal.ReceiverSide,
				},
				Meta: conn.Meta,
			},
			emitterNodeName: emitterNodeName,
			emitterNode:     emitterNode,
		},
	}, nil
}

func (d Desugarer) handleConstRefSender(
	conn src.Connection,
	scope src.Scope,
) (
	handleConstRefSenderResult,
	*compiler.Error,
) {
	constTypeExpr, err := d.getConstTypeByRef(*conn.Normal.SenderSide.Const.Ref, scope)
	if err != nil {
		return handleConstRefSenderResult{}, compiler.Error{
			Err: fmt.Errorf(
				"Unable to get constant type by reference '%v'",
				*conn.Normal.SenderSide.Const.Ref,
			),
			Location: &scope.Location,
			Meta:     &conn.Normal.SenderSide.Const.Ref.Meta,
		}.Wrap(err)
	}

	counter := virtualEmittersCount.Load()
	virtualEmittersCount.Store(counter + 1)
	virtualEmitterName := fmt.Sprintf("virtual_emitter_%d", counter)

	emitterNode := src.Node{
		Directives: map[src.Directive][]string{
			compiler.BindDirective: {
				conn.Normal.SenderSide.Const.Ref.String(), // don't forget to bind const
			},
		},
		EntityRef: emitterComponentRef,
		TypeArgs:  []ts.Expr{constTypeExpr},
	}
	emitterNodeOutportAddr := src.PortAddr{
		Node: virtualEmitterName,
		Port: "msg",
	}

	return handleConstRefSenderResult{
		connectionWithoutConstSender: src.Connection{
			Normal: &src.NormalConnection{
				SenderSide: src.ConnectionSenderSide{
					PortAddr:  &emitterNodeOutportAddr,
					Selectors: conn.Normal.SenderSide.Selectors,
					Meta:      conn.Normal.SenderSide.Meta,
				},
				ReceiverSide: conn.Normal.ReceiverSide,
			},
			Meta: conn.Meta,
		},
		emitterNodeName: virtualEmitterName,
		emitterNode:     emitterNode,
	}, nil
}

// getConstTypeByRef is needed to figure out type parameters for Const node
func (d Desugarer) getConstTypeByRef(ref core.EntityRef, scope src.Scope) (ts.Expr, *compiler.Error) {
	entity, _, err := scope.Entity(ref)
	if err != nil {
		return ts.Expr{}, &compiler.Error{
			Err:      err,
			Location: &scope.Location,
			Meta:     &ref.Meta,
		}
	}

	if entity.Kind != src.ConstEntity {
		return ts.Expr{}, &compiler.Error{
			Err:      fmt.Errorf("%w: %v", ErrConstSenderEntityKind, entity.Kind),
			Location: &scope.Location,
			Meta:     entity.Meta(),
		}
	}

	if entity.Const.Ref != nil {
		expr, err := d.getConstTypeByRef(*entity.Const.Ref, scope)
		if err != nil {
			return ts.Expr{}, compiler.Error{
				Location: &scope.Location,
				Meta:     entity.Meta(),
			}.Wrap(err)
		}
		return expr, nil
	}

	return entity.Const.Message.TypeExpr, nil
}
