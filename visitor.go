/* Package statemachine
 * @Author 砚池/Ivan
 * @Date 2024/04/09
 * @Description:
 */

package statemachine

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Visitable[S, E comparable, C any] interface {
	Accept(visitor VisitorInterface[S, E, C]) string
}

type VisitorInterface[S, E comparable, C any] interface {
	visitOnStatemachineEntry(s *StateMachine[S, E, C]) string
	visitOnStatemachineExit(s *StateMachine[S, E, C]) string
	visitOnStateEntry(s *state[S, E, C]) string
	visitOnStateExit(s *state[S, E, C]) string
}

type SysOutVisitor[S, E comparable, C any] struct{}

func (v *SysOutVisitor[S, E, C]) visitOnStatemachineEntry(s *StateMachine[S, E, C]) string {
	entry := fmt.Sprintf("-----StateMachine:%s-------", s.machineId)
	log.Println(entry)
	return entry
}

func (v *SysOutVisitor[S, E, C]) visitOnStatemachineExit(s *StateMachine[S, E, C]) string {
	exit := "------------------------"
	log.Println(exit)
	return exit
}

func (v *SysOutVisitor[S, E, C]) visitOnStateEntry(s *state[S, E, C]) string {
	sb := strings.Builder{}
	stateStr := "State:" + s.String()
	log.Println(stateStr)
	sb.WriteString(stateStr + "\n")
	for _, t := range s.allTransitions() {
		tStr := "    Transition:" + t.String()
		log.Println(tStr)
		sb.WriteString(tStr + "\n")
	}
	return sb.String()
}

func (v *SysOutVisitor[S, E, C]) visitOnStateExit(s *state[S, E, C]) string {
	return ""
}
