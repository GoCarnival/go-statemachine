# go-statemachine

参考了[alibaba/COLA/cola-component-statemachine](https://github.com/alibaba/COLA/tree/master/cola-components/cola-component-statemachine)
的实现

## Install

```shell
go get github.com/GoCarnival/go-statemachine
```

## Usage

```go
package main

import (
	"github.com/GoCarnival/go-statemachine"
	"fmt"
)

func main() {
	builder := statemachine.Builder[string, string, any]{}
	builder.ExternalTransition().From("foo").To("bar").On("ping").When(func(ctx any) bool {
		return true
	}).Perform(func(from string, to string, event string, context any) {
		fmt.Println("do something")
	})
	var fetcher statemachine.CurrentStateFetcher[string, any] = func(ctx any) string {
		return "foo"
	}
	builder.SetCurrentStateFetcher(fetcher)
	s := builder.Build("test")
	target := s.FireEvent("foo", "ping", nil)
	fmt.Println(target)
	s.Verify("foo", "ping")
	s.FireEventByFetcher("ping", nil)
	s.VerifyWithFetcher("ping", nil)
}

```