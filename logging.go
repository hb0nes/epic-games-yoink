package main

import (
	"context"
	"fmt"

	l "github.com/chromedp/cdproto/log"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func setupLogger(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if e, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				for _, arg := range e.Args {
					if arg.Type != runtime.TypeUndefined {
						// if e.Type == runtime.APITypeError && arg.Type != runtime.TypeUndefined {
						fmt.Printf("Console Entry: %s\n", arg.Value)
					}
				}
			}
			if e, ok := ev.(*l.EventEntryAdded); ok {
				fmt.Printf("Console Log Entry: %s\n", e.Entry.Text)
			}
		}()
	})
}
