/* Package statemachine
 * @Author 砚池/Ivan
 * @Date 2024/04/08
 * @Description:
 */

package statemachine

import (
	log "github.com/sirupsen/logrus"
	"strings"
)

// StateMachine the state machine
type StateMachine[S, E comparable, C any] struct {
	machineId           string
	stateMap            map[S]state[S, E, C]
	ready               bool
	failCallback        FailCallbackInterface[S, E, C]
	currentStateFetcher CurrentStateFetcher[S, C]
}

func (s *StateMachine[S, E, C]) isReady() {
	if !s.ready {
		log.Fatal("state machine is not built yet, can not work")
	}
}

// Verify 校验状态机是否可以触发某个事件
func (s *StateMachine[S, E, C]) Verify(source S, event E) bool {
	s.isReady()
	sourceState := getState(s.stateMap, source)
	transitions := sourceState.transitionsOfEvent(event)
	return transitions != nil && len(transitions) != 0
}

// VerifyWithFetcher 校验状态机是否可以触发某个事件，当前状态通过 CurrentStateFetcher 获取，需要先在 StateMachine 上配置 CurrentStateFetcher
func (s *StateMachine[S, E, C]) VerifyWithFetcher(event E, ctx C) bool {
	s.isReady()
	if s.currentStateFetcher == nil {
		log.Fatal("no state fetcher configured")
	}
	source := s.currentStateFetcher(ctx)
	sourceState := getState(s.stateMap, source)
	transitions := sourceState.transitionsOfEvent(event)
	return transitions != nil && len(transitions) != 0
}

func (s *StateMachine[S, E, C]) routeTransition(source S, event E, ctx C) *transition[S, E, C] {
	sourceState := getState(s.stateMap, source)
	transitions := sourceState.transitionsOfEvent(event)
	if transitions == nil || len(transitions) == 0 {
		return nil
	}
	var transit *transition[S, E, C]
	for _, transition := range transitions {
		if transition.condition == nil {
			transit = transition
		} else if transition.condition(ctx) {
			transit = transition
			break
		}
	}
	return transit
}

// FireEvent 触发状态机事件
func (s *StateMachine[S, E, C]) FireEvent(source S, event E, ctx C) (target S) {
	s.isReady()
	transition := s.routeTransition(source, event, ctx)
	if transition == nil {
		s.failCallback.OnFail(source, event, ctx)
		return source
	}
	return transition.transit(ctx, false).id
}

// FireEventByFetcher 触发状态机事件，当前状态通过 CurrentStateFetcher 获取，需要先在 StateMachine 上配置 CurrentStateFetcher
func (s *StateMachine[S, E, C]) FireEventByFetcher(event E, ctx C) (target S) {
	s.isReady()
	if s.currentStateFetcher == nil {
		log.Fatal("no state fetcher configured")
	}
	source := s.currentStateFetcher(ctx)
	transition := s.routeTransition(source, event, ctx)
	if transition == nil {
		s.failCallback.OnFail(source, event, ctx)
		return source
	}
	return transition.transit(ctx, false).id
}

func (s *StateMachine[S, E, C]) Accept(v VisitorInterface[S, E, C]) string {
	sb := strings.Builder{}
	sb.WriteString(v.visitOnStatemachineEntry(s))
	for _, state := range s.stateMap {
		sb.WriteString(v.visitOnStateEntry(&state))
	}
	sb.WriteString(v.visitOnStatemachineExit(s))
	return sb.String()
}

func (s *StateMachine[S, E, C]) ShowStateMachine() {
	s.Accept(&SysOutVisitor[S, E, C]{})
}

type Builder[S, E comparable, C any] struct {
	stateMap            map[S]state[S, E, C]
	statemachine        *StateMachine[S, E, C]
	failCallback        FailCallbackInterface[S, E, C]
	currentStateFetcher CurrentStateFetcher[S, C]
}

// SetFailCallback 设置状态转移失败后的回调
func (b *Builder[S, E, C]) SetFailCallback(c FailCallbackInterface[S, E, C]) {
	b.failCallback = c
}

// SetCurrentStateFetcher 设置 CurrentStateFetcher，用于获取当前 state
func (b *Builder[S, E, C]) SetCurrentStateFetcher(f CurrentStateFetcher[S, C]) {
	b.currentStateFetcher = f
}

// ExternalTransition 外部状态转移，A->B
func (b *Builder[S, E, C]) ExternalTransition() ExternalTransitionInterface[S, E, C] {
	if b.stateMap == nil {
		b.stateMap = make(map[S]state[S, E, C])
	}
	return newExternalTransitionBuilder(b.stateMap, EXTERNAL)
}

// ExternalTransitions 外部状态转移，（A，B）-> C
func (b *Builder[S, E, C]) ExternalTransitions() ExternalTransitionsInterface[S, E, C] {
	if b.stateMap == nil {
		b.stateMap = make(map[S]state[S, E, C])
	}
	return newExternalTransitionsBuilder(b.stateMap, EXTERNAL)
}

// InternalTransition 内部状态转移，也可以理解为没有状态转换，A->A，用于触发在某个状态响应某个事件，执行某些逻辑
func (b *Builder[S, E, C]) InternalTransition() InternalTransitionInterface[S, E, C] {
	if b.stateMap == nil {
		b.stateMap = make(map[S]state[S, E, C])
	}
	return newInternalTransitionBuilder(b.stateMap, INTERNAL)
}

// Build 构建状态机
func (b *Builder[S, E, C]) Build(machineId string) *StateMachine[S, E, C] {
	if b.statemachine == nil {
		b.statemachine = &StateMachine[S, E, C]{}
	}
	statemachine := b.statemachine
	statemachine.stateMap = b.stateMap
	statemachine.machineId = machineId
	statemachine.ready = true
	statemachine.currentStateFetcher = b.currentStateFetcher
	if b.failCallback == nil {
		statemachine.failCallback = &NumbFailCallback[S, E, C]{}
	} else {
		statemachine.failCallback = b.failCallback
	}
	return statemachine
}
