package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
)

const (
	MouseLeft uint16 = iota + 1
	MouseRight
	MouseMiddle
)

type multiplexer struct {
	i   int
	sub map[int]chan [3]string
	mu  sync.Mutex
}

func newWatcher() {
	log.Println("main.newWatcher()")
	multi = multiplexer{
		i:   0,
		sub: map[int]chan [3]string{},
	}

	evCh := hook.Start()
	defer hook.End()

	const mouseMoveRateLimit = 30 * time.Millisecond // 30ms => 約30fps
	var lastMouseMove time.Time

	var buf [3]string
	for ev := range evCh {
		switch ev.Kind {
		case hook.KeyDown:
			buf = [3]string{"keyboard", "down", hook.RawcodetoKeychar(ev.Rawcode)}
		case hook.KeyUp:
			buf = [3]string{"keyboard", "up", hook.RawcodetoKeychar(ev.Rawcode)}
		case hook.MouseDown:
			buf = [3]string{"mouse", "down", ""}
			switch ev.Button {
			case MouseLeft:
				buf[2] = "Left"
			case MouseMiddle:
				buf[2] = "Middle"
			case MouseRight:
				buf[2] = "Right"
			}
		case hook.MouseUp:
			buf = [3]string{"mouse", "up", ""}
			switch ev.Button {
			case MouseLeft:
				buf[2] = "Left"
			case MouseMiddle:
				buf[2] = "Middle"
			case MouseRight:
				buf[2] = "Right"
			}
		case hook.MouseMove:
			now := time.Now()
			if now.Sub(lastMouseMove) < mouseMoveRateLimit {
				continue
			}

			lastMouseMove = now
			buf = [3]string{"mouse", "move", fmt.Sprintf("[%d,%d]", ev.X, ev.Y)}
		}
		if *debug {
			log.Println(buf)
		}

		for _, s := range multi.sub {
			select {
			case s <- buf:
			default:
			}
		}
	}
}
