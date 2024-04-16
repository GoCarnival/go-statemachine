/* Package statemachine
 * @Author 砚池/Ivan
 * @Date 2024/04/08
 * @Description:
 */

package statemachine

import (
	"fmt"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

// Action 是状态机规则命中后，执行的业务逻辑
type Action[S, E comparable, C any] func(from S, to S, event E, context *C)

// Condition 是状态机路由后在执行 Action 之前的前置校验，在配置状态机时可以为空
type Condition[C any] func(ctx *C) bool

// TransitionType TransitionType 状态转移类型
// 当前只有 EXTERNAL 与 INTERNAL 两种
type TransitionType string

const (
	//INTERNAL  if triggered, occurs without exiting or entering the source state
	//(i.e., it does not cause a state change). This means that the entry or exit condition of the source
	//state will not be invoked. An internal transition can be taken even if the SateMachine is in one or
	//more Regions nested within the associated state.
	INTERNAL TransitionType = "INTERNAL"
	LOCAL    TransitionType = "LOCAL"
	//EXTERNAL if triggered, will exit the composite (source) state.
	EXTERNAL TransitionType = "EXTERNAL"
)

// transition is something what a state machine associates with a state
type transition[S, E comparable, C any] struct {
	source    state[S, E, C]
	target    state[S, E, C]
	event     E
	condition Condition[C]
	t         TransitionType
	action    Action[S, E, C]
}

// transit Do transition from source state to target state.
func (t *transition[S, E, C]) transit(ctx *C, checkCondition bool) state[S, E, C] {
	log.Debugf("Do t: %s\n", t.String())
	err := t.verify()
	if err != nil {
		log.Errorf("err:%s,stay at the %v state\n", err.Error(), t.source)
		return t.source
	}
	if !checkCondition || t.condition == nil || t.condition(ctx) {
		if t.action != nil {
			t.action(t.source.id, t.target.id, t.event, ctx)
		}
		return t.target
	}
	log.Debugf("Condition is not satisfied, stay at the %v state\n", t.source)
	return t.source
}

// verify transition correctness
func (t *transition[S, E, C]) verify() error {
	if t.t == INTERNAL && t.source.id != t.target.id {
		return fmt.Errorf("Internal t source state '%v' and target state '%v' must be same.\n", t.source.id, t.target.id)
	}
	return nil
}

// String 打印状态机
func (t *transition[S, E, C]) String() string {
	return fmt.Sprintf("%v-[%v+%v]->%v", t.source.id, t.event, t.t, t.target.id)
}

func sameTransition[S, E comparable, C any](first, second *transition[S, E, C]) bool {
	fs := first.source.id
	ft := first.target.id
	fe := first.event
	ss := second.source.id
	st := second.target.id
	se := second.event
	return fs == ss && ft == st && fe == se
}

// state 状态
type state[S, E comparable, C any] struct {
	id               S
	eventTransitions *eventTransitions[S, E, C]
}

func (s *state[S, E, C]) String() string {
	return fmt.Sprintf("%v", s.id)
}

// addTransition to the state
func (s *state[S, E, C]) addTransition(event E, target state[S, E, C], transitionType TransitionType) *transition[S, E, C] {
	t := &transition[S, E, C]{
		source: *s,
		target: target,
		event:  event,
		t:      transitionType,
	}
	s.transitions().put(event, t)
	return t
}

func (s *state[S, E, C]) transitions() *eventTransitions[S, E, C] {
	if s.eventTransitions == nil {
		s.eventTransitions = &eventTransitions[S, E, C]{}
	}
	return s.eventTransitions
}

func (s *state[S, E, C]) transitionsOfEvent(event E) []*transition[S, E, C] {
	return s.transitions().get(event)
}

func (s *state[S, E, C]) allTransitions() []*transition[S, E, C] {
	return s.transitions().allTransitions()
}

type CurrentStateFetcher[S comparable, C any] func(ctx *C) S

type eventTransitions[S, E comparable, C any] struct {
	transitions map[E][]*transition[S, E, C]
}

func (et *eventTransitions[S, E, C]) put(event E, t *transition[S, E, C]) {
	if et.transitions == nil {
		et.transitions = make(map[E][]*transition[S, E, C])
	}
	if et.transitions[event] == nil {
		var transitions []*transition[S, E, C]
		transitions = append(transitions, t)
		et.transitions[event] = transitions
	} else {
		existingTransitions := et.transitions[event]
		err := et.verify(existingTransitions, t)
		if err != nil {
			log.Fatal(err)
		}
		existingTransitions = append(existingTransitions, t)
		et.transitions[event] = existingTransitions
	}
}

func (et *eventTransitions[S, E, C]) get(event E) []*transition[S, E, C] {
	return et.transitions[event]
}

func (et *eventTransitions[S, E, C]) allTransitions() []*transition[S, E, C] {
	var all []*transition[S, E, C]
	for _, transitions := range et.transitions {
		all = append(all, transitions...)
	}
	return all
}

func (et *eventTransitions[S, E, C]) verify(existingTransitions []*transition[S, E, C], newTransition *transition[S, E, C]) error {
	for _, transition := range existingTransitions {
		if sameTransition(transition, newTransition) {
			return fmt.Errorf("%v already Exist, you can not add another one\n", transition)
		}
	}
	return nil
}

type fromInterface[S, E comparable, C any] interface {
	To(stateId S) toInterface[S, E, C]
}

type toInterface[S, E comparable, C any] interface {
	On(event E) onInterface[S, E, C]
}

type onInterface[S, E comparable, C any] interface {
	When(condition Condition[C]) whenInterface[S, E, C]
	whenInterface[S, E, C]
}

type whenInterface[S, E comparable, C any] interface {
	Perform(action Action[S, E, C])
}

type ExternalTransitionInterface[S, E comparable, C any] interface {
	From(stateId S) fromInterface[S, E, C]
}

type ExternalTransitionsInterface[S, E comparable, C any] interface {
	FromAmong(stateIds ...S) fromInterface[S, E, C]
}

type InternalTransitionInterface[S, E comparable, C any] interface {
	Within(stateId S) toInterface[S, E, C]
}

type transitionBuilder[S, E comparable, C any] struct {
	stateMap   map[S]state[S, E, C]
	target     state[S, E, C]
	t          TransitionType
	source     state[S, E, C]
	transition *transition[S, E, C]
}

func newExternalTransitionBuilder[S, E comparable, C any](stateMap map[S]state[S, E, C], t TransitionType) ExternalTransitionInterface[S, E, C] {
	return &transitionBuilder[S, E, C]{
		stateMap: stateMap,
		t:        t,
	}
}

func newInternalTransitionBuilder[S, E comparable, C any](stateMap map[S]state[S, E, C], t TransitionType) InternalTransitionInterface[S, E, C] {
	return &transitionBuilder[S, E, C]{
		stateMap: stateMap,
		t:        t,
	}
}

func (tb *transitionBuilder[S, E, C]) Perform(action Action[S, E, C]) {
	tb.transition.action = action
}

func (tb *transitionBuilder[S, E, C]) On(event E) onInterface[S, E, C] {
	tb.transition = tb.source.addTransition(event, tb.target, tb.t)
	return tb
}

func (tb *transitionBuilder[S, E, C]) When(condition Condition[C]) whenInterface[S, E, C] {
	tb.transition.condition = condition
	return tb
}

func (tb *transitionBuilder[S, E, C]) To(stateId S) toInterface[S, E, C] {
	tb.target = getState(tb.stateMap, stateId)
	return tb
}

func (tb *transitionBuilder[S, E, C]) From(stateId S) fromInterface[S, E, C] {
	tb.source = getState(tb.stateMap, stateId)
	return tb
}

func (tb *transitionBuilder[S, E, C]) Within(stateId S) toInterface[S, E, C] {
	tb.source = getState(tb.stateMap, stateId)
	tb.target = getState(tb.stateMap, stateId)
	return tb
}

func getState[S, E comparable, C any](stateMap map[S]state[S, E, C], stateId S) state[S, E, C] {
	if !lo.Contains(lo.Keys(stateMap), stateId) {
		s := state[S, E, C]{
			id:               stateId,
			eventTransitions: &eventTransitions[S, E, C]{},
		}
		stateMap[stateId] = s
	}
	return stateMap[stateId]
}

type transitionsBuilder[S, E comparable, C any] struct {
	stateMap    map[S]state[S, E, C]
	target      state[S, E, C]
	t           TransitionType
	sources     []state[S, E, C]
	transitions []*transition[S, E, C]
}

func newExternalTransitionsBuilder[S, E comparable, C any](stateMap map[S]state[S, E, C], t TransitionType) ExternalTransitionsInterface[S, E, C] {
	return &transitionsBuilder[S, E, C]{
		stateMap: stateMap,
		t:        t,
	}
}

func (tb *transitionsBuilder[S, E, C]) To(stateId S) toInterface[S, E, C] {
	tb.target = getState(tb.stateMap, stateId)
	return tb
}

func (tb *transitionsBuilder[S, E, C]) On(event E) onInterface[S, E, C] {
	transitions := tb.transitions
	for _, source := range tb.sources {
		transition := source.addTransition(event, tb.target, tb.t)
		transitions = append(transitions, transition)
	}
	tb.transitions = transitions
	return tb
}

func (tb *transitionsBuilder[S, E, C]) FromAmong(stateIds ...S) fromInterface[S, E, C] {
	sources := tb.sources
	for _, stateId := range stateIds {
		sources = append(sources, getState(tb.stateMap, stateId))
	}
	tb.sources = sources
	return tb
}

func (tb *transitionsBuilder[S, E, C]) Perform(action Action[S, E, C]) {
	for _, transition := range tb.transitions {
		transition.action = action
	}
}

func (tb *transitionsBuilder[S, E, C]) When(condition Condition[C]) whenInterface[S, E, C] {
	for _, transition := range tb.transitions {
		transition.condition = condition
	}
	return tb
}

type FailCallbackInterface[S, E comparable, C any] interface {
	OnFail(source S, event E, ctx *C)
}
type NumbFailCallback[S, E comparable, C any] struct{}

func (n *NumbFailCallback[S, E, C]) OnFail(source S, event E, ctx *C) {
	//do nothing
}

type AlertFailCallback[S, E comparable, C any] struct{}

func (a *AlertFailCallback[S, E, C]) OnFail(source S, event E, ctx *C) {
	log.Fatalf("Cannot fire event [%v] on current state [%v] with context [%v]\n", event, source, ctx)
}
