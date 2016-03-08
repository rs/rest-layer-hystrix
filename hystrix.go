// Package restrix is a REST Layer resource storage wrapper to add hystrix support
// to the underlaying storage handler.
package restrix

import (
	"fmt"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/rs/rest-layer/resource"
	"golang.org/x/net/context"
)

type wrapper struct {
	resource.Storer
	getCmd    string
	findCmd   string
	insertCmd string
	updateCmd string
	deleteCmd string
	clearCmd  string
}

type mgetWrapper struct {
	wrapper
	multiGetCmd string
}

// Wrap wraps a REST Layer storage handler to add hystrix support to all
// handler's methods.
//
// Hystrix wraps each storage handlers into an hystrix command. One hystrix
// command is created per backend actions with the format <name>.<Action>.
//
// Actions are Find, Insert, Update, Delete, Clear and MultiGet for handlers
// implementing MultiGetter interface.
//
// You must configure hystrix for each command you want to control and start the
// stream handler.
// See https://godoc.org/github.com/afex/hystrix-go/hystrix for more info and
// examples/hystrix/main.go for a usage example.
func Wrap(name string, h resource.Storer) resource.Storer {
	w := wrapper{
		Storer:    h,
		getCmd:    fmt.Sprintf("%s.Get", name),
		findCmd:   fmt.Sprintf("%s.Find", name),
		insertCmd: fmt.Sprintf("%s.Insert", name),
		updateCmd: fmt.Sprintf("%s.Update", name),
		deleteCmd: fmt.Sprintf("%s.Delete", name),
		clearCmd:  fmt.Sprintf("%s.Clear", name),
	}
	if _, ok := h.(resource.MultiGetter); ok {
		return mgetWrapper{
			wrapper:     w,
			multiGetCmd: fmt.Sprintf("%s.MultiGet", name),
		}
	}
	return w
}

func (w wrapper) Insert(ctx context.Context, items []*resource.Item) error {
	return hystrix.Do(w.insertCmd, func() error {
		return w.Storer.Insert(ctx, items)
	}, nil)
}

func (w wrapper) Update(ctx context.Context, item *resource.Item, original *resource.Item) error {
	return hystrix.Do(w.updateCmd, func() error {
		return w.Storer.Update(ctx, item, original)
	}, nil)
}

func (w wrapper) Delete(ctx context.Context, item *resource.Item) error {
	return hystrix.Do(w.deleteCmd, func() error {
		return w.Storer.Delete(ctx, item)
	}, nil)
}

func (w wrapper) Clear(ctx context.Context, lookup *resource.Lookup) (deleted int, err error) {
	out := make(chan int, 1)
	errs := hystrix.Go(w.clearCmd, func() error {
		deleted, err := w.Storer.Clear(ctx, lookup)
		if err == nil {
			out <- deleted
		}
		return err
	}, nil)
	select {
	case deleted = <-out:
	case err = <-errs:
	}
	return
}

func (w wrapper) Find(ctx context.Context, lookup *resource.Lookup, page, perPage int) (list *resource.ItemList, err error) {
	out := make(chan *resource.ItemList, 1)
	errs := hystrix.Go(w.findCmd, func() error {
		list, err := w.Storer.Find(ctx, lookup, page, perPage)
		if err == nil {
			out <- list
		}
		return err
	}, nil)
	select {
	case list = <-out:
	case err = <-errs:
	}
	return
}

func (w mgetWrapper) MultiGet(ctx context.Context, ids []interface{}) (items []*resource.Item, err error) {
	out := make(chan []*resource.Item, 1)
	errs := hystrix.Go(w.multiGetCmd, func() error {
		mg := w.wrapper.Storer.(resource.MultiGetter)
		items, err := mg.MultiGet(ctx, ids)
		if err == nil {
			out <- items
		}
		return err
	}, nil)
	select {
	case items = <-out:
	case err = <-errs:
	}
	return
}
