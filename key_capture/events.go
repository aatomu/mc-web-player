package main

type CaptureEvent struct {
	mouse    MouseEvent
	keyboard KeyboardEvent
}

type MouseEvent struct {
	move    chan [2]int32
	down    chan string
	up      chan string
	Vscroll chan int16
	Hscroll chan int16
}

type KeyboardEvent struct {
	down chan string
	up   chan string
}
