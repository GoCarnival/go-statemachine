/* Package statemachine
 * @Author 砚池/Ivan
 * @Date 2024/04/08
 * @Description:
 */

package statemachine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	f            Factory[string, string, any]
	statemachine *StateMachine[string, string, any]
)

func init() {
	f = Factory[string, string, any]{}

	b := Builder[string, string, any]{}
	b.ExternalTransition().From("bar").To("baz").On("hi").Perform(func(from string, to string, event string, context *any) {
		fmt.Printf("from %v ,to %v,on %v\n", from, to, event)
	})
	b.ExternalTransition().From("foo").To("bar").On("ping").When(func(ctx *any) bool {
		fmt.Println("check true")
		return true
	}).Perform(func(from string, to string, event string, context *any) {
		fmt.Printf("from %v ,to %v,on %v\n", from, to, event)
	})
	b.ExternalTransitions().FromAmong("foo", "bar").To("zzz").On("sleep").When(func(ctx *any) bool {
		return true
	}).Perform(func(from string, to string, event string, context *any) {
		fmt.Printf("from %v ,to %v,on %v\n", from, to, event)
	})
	b.InternalTransition().Within("foo").On("in").Perform(func(from string, to string, event string, context *any) {
		fmt.Printf("from %v ,to %v,on %v\n", from, to, event)
	})
	var fetcher CurrentStateFetcher[string, any] = func(ctx *any) string {
		return "foo"
	}
	b.SetCurrentStateFetcher(fetcher)
	b.SetFailCallback(&NumbFailCallback[string, string, any]{})
	statemachine = b.Build("test")
	_ = f.Register(statemachine)
}

func TestShowStatemachine(t *testing.T) {
	statemachine.ShowStateMachine()
}

func TestFireEvent(t *testing.T) {
	is := assert.New(t)
	target := statemachine.FireEvent("foo", "ping", nil)
	is.Equal("bar", target)
	target = statemachine.FireEvent("bar", "zzz", nil)
	is.Equal("bar", target)
	target = statemachine.FireEvent("bar", "hi", nil)
	is.Equal("baz", target)
	target = statemachine.FireEvent("foo", "in", nil)
	is.Equal("foo", target)
	target = statemachine.FireEvent("foo", "sleep", nil)
	is.Equal("zzz", target)
	target = statemachine.FireEvent("bar", "sleep", nil)
	is.Equal("zzz", target)
	target = statemachine.FireEventByFetcher("ping", nil)
	is.Equal("bar", target)
}

func TestVerify(t *testing.T) {
	is := assert.New(t)
	v := statemachine.Verify("foo", "ping")
	is.True(v)
	v = statemachine.Verify("foo", "xxx")
	is.False(v)
	v = statemachine.VerifyWithFetcher("ping", nil)
	is.True(v)
	v = statemachine.VerifyWithFetcher("xxx", nil)
	is.False(v)
}

func TestFactory(t *testing.T) {
	sm, err := f.Get("test")
	if err != nil {
		panic(err)
	}
	sm.ShowStateMachine()
}
