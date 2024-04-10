/* Package statemachine
 * @Author 砚池/Ivan
 * @Date 2024/04/09
 * @Description:
 */

package statemachine

import (
	"fmt"
	"github.com/samber/lo"
)

type Factory[S, E comparable, C any] struct {
	container map[string]*StateMachine[S, E, C]
}

func (f *Factory[S, E, C]) Register(s *StateMachine[S, E, C]) error {
	machineId := s.machineId
	if f.container == nil {
		f.container = make(map[string]*StateMachine[S, E, C])
	}
	if lo.Contains(lo.Keys(f.container), machineId) {
		return fmt.Errorf("the state machine with id [%s] is already built, no need to build again\n", machineId)
	}
	f.container[machineId] = s
	return nil
}

func (f *Factory[S, E, C]) Get(machineId string) (*StateMachine[S, E, C], error) {
	if !lo.Contains(lo.Keys(f.container), machineId) {
		return nil, fmt.Errorf("there is no stateMachine instance for %s, please build it first\n", machineId)
	}
	s := f.container[machineId]
	return s, nil
}
