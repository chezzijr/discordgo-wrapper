package main

import "errors"

// Base error
var (
	ErrTimeout 			= errors.New("err: Timeout")
)

// Error for pagination and widget
var (
	ErrAlreadyRunning 	= errors.New("err: Widget already running")
	ErrIndexOutOfBounds = errors.New("err: Index is out of bounds")
	ErrNilMessage       = errors.New("err: Message is nil")
	ErrNilPage         	= errors.New("err: MessageSend is nil")
	ErrNotRunning       = errors.New("err: Not running")
	ErrPagesEmpty		= errors.New("err: No page")
)