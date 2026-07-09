package interactions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

var ErrRouteNotFound = errors.New("interaction route not found")

type Handler func(ctx context.Context, interaction Interaction, responder responses.Responder) error

type Middleware func(next Handler) Handler

type Router struct {
	slash        map[string]Handler
	components   map[RouteKey]Handler
	modals       map[RouteKey]Handler
	autocomplete map[string]Handler
	middleware   []Middleware
	idParser     CustomIDParser
}

func NewRouter(middleware ...Middleware) *Router {
	return &Router{
		slash:        map[string]Handler{},
		components:   map[RouteKey]Handler{},
		modals:       map[RouteKey]Handler{},
		autocomplete: map[string]Handler{},
		middleware:   append([]Middleware(nil), middleware...),
	}
}

func (r *Router) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *Router) SetCustomIDParser(parser CustomIDParser) {
	r.idParser = parser
}

func (r *Router) RegisterSlash(name string, handler Handler) error {
	if name == "" {
		return errors.New("slash route name is required")
	}
	if handler == nil {
		return errors.New("slash route handler is required")
	}
	r.slash[name] = handler
	return nil
}

func (r *Router) RegisterComponent(key ComponentKey, handler Handler) error {
	if key.Version == "" || key.Feature == "" || key.Action == "" {
		return errors.New("component route key is incomplete")
	}
	if handler == nil {
		return errors.New("component route handler is required")
	}
	r.components[RouteKey{Kind: TypeComponent, Version: key.Version, Feature: key.Feature, Action: key.Action}] = handler
	return nil
}

func (r *Router) RegisterModal(key ModalKey, handler Handler) error {
	if key.Version == "" || key.Feature == "" || key.Action == "" {
		return errors.New("modal route key is incomplete")
	}
	if handler == nil {
		return errors.New("modal route handler is required")
	}
	r.modals[RouteKey{Kind: TypeModal, Version: key.Version, Feature: key.Feature, Action: key.Action}] = handler
	return nil
}

func (r *Router) RegisterRoute(key RouteKey, handler Handler) error {
	if key.Kind != TypeComponent && key.Kind != TypeModal {
		return errors.New("route key kind must be component or modal")
	}
	if key.Version == "" || key.Feature == "" || key.Action == "" {
		return errors.New("route key is incomplete")
	}
	if handler == nil {
		return errors.New("route handler is required")
	}
	if key.Kind == TypeComponent {
		r.components[key] = handler
		return nil
	}
	r.modals[key] = handler
	return nil
}

func (r *Router) RegisterAutocomplete(name string, handler Handler) error {
	if name == "" {
		return errors.New("autocomplete route name is required")
	}
	if handler == nil {
		return errors.New("autocomplete route handler is required")
	}
	r.autocomplete[name] = handler
	return nil
}

func (r *Router) Handle(ctx context.Context, interaction Interaction, responder responses.Responder) error {
	var err error
	if r.idParser != nil {
		interaction, err = ApplyParsedRoute(interaction, r.idParser, interaction.ModalFields)
		if err != nil {
			if responder != nil {
				_ = responder.Error(ctx, err)
			}
			return err
		}
	}
	handler, err := r.lookup(interaction)
	if err != nil {
		return err
	}
	return Chain(handler, r.middleware...)(ctx, interaction, responder)
}

func (r *Router) lookup(interaction Interaction) (Handler, error) {
	switch interaction.Type {
	case TypeSlash:
		if handler, ok := r.slash[interaction.CommandName]; ok {
			return handler, nil
		}
	case TypeComponent:
		key := interaction.RouteKey
		if key.IsZero() {
			key = RouteKey{Kind: TypeComponent, Version: interaction.ComponentKey.Version, Feature: interaction.ComponentKey.Feature, Action: interaction.ComponentKey.Action}
		}
		if handler, ok := r.components[key]; ok {
			return handler, nil
		}
	case TypeModal:
		key := interaction.RouteKey
		if key.IsZero() {
			key = RouteKey{Kind: TypeModal, Version: interaction.ModalKey.Version, Feature: interaction.ModalKey.Feature, Action: interaction.ModalKey.Action}
		}
		if handler, ok := r.modals[key]; ok {
			return handler, nil
		}
	case TypeAutocomplete:
		if handler, ok := r.autocomplete[interaction.CommandName]; ok {
			return handler, nil
		}
	default:
		return nil, fmt.Errorf("%w: unsupported interaction type %q", ErrRouteNotFound, interaction.Type)
	}
	return nil, fmt.Errorf("%w: %s", ErrRouteNotFound, interaction.Route().String())
}

func Chain(handler Handler, middleware ...Middleware) Handler {
	wrapped := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	return wrapped
}
