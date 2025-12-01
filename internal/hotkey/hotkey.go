package hotkey

import (
	"context"
	"fmt"

	"golang.design/x/hotkey"
)

type Listener struct {
	hk     *hotkey.Hotkey
	ctx    context.Context
	cancel context.CancelFunc
	onDown func()
	onUp   func()
}

func NewListener(parent context.Context, onDown, onUp func()) (*Listener, error) {
	ctx, cancel := context.WithCancel(parent)
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyF4)
	if err := hk.Register(); err != nil {
		cancel()
		return nil, fmt.Errorf("register hotkey: %w", err)
	}
	l := &Listener{
		hk:     hk,
		ctx:    ctx,
		cancel: cancel,
		onDown: onDown,
		onUp:   onUp,
	}
	go l.loop()
	return l, nil
}

func (l *Listener) loop() {
	defer l.hk.Unregister()
	for {
		select {
		case <-l.ctx.Done():
			return
		case <-l.hk.Keydown():
			if l.onDown != nil {
				l.onDown()
			}
		case <-l.hk.Keyup():
			if l.onUp != nil {
				l.onUp()
			}
		}
	}
}

func (l *Listener) Close() {
	l.cancel()
}
