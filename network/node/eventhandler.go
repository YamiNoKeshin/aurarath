package node

import (
	"github.com/hashicorp/serf/serf"
	"github.com/joernweissenborn/eventual2go"
)

func newEventHandler() (eh eventHandler) {
	eh.stream = eventual2go.NewStreamController()
	eh.join = eh.stream.Where(isJoin).Transform(toMember)
	eh.leave = eh.stream.Where(isLeave).Transform(toMember)
	eh.query = eh.stream.Where(isQuery)
	return
}

type eventHandler struct {
	stream *eventual2go.StreamController
	join   *eventual2go.Stream
	leave  *eventual2go.Stream
	query  *eventual2go.Stream
}

func (eh eventHandler) HandleEvent(evt serf.Event) {
	eh.stream.Add(evt)
}

func isJoin(d eventual2go.Data) (is bool) {
	return d.(serf.Event).EventType() == serf.EventMemberJoin
}

func isLeave(d eventual2go.Data) (is bool) {
	return d.(serf.Event).EventType() == serf.EventMemberLeave || d.(serf.Event).EventType() == serf.EventMemberFailed
}

func isQuery(d eventual2go.Data) (is bool) {
	return d.(serf.Event).EventType() == serf.EventQuery
}

func toMember(d eventual2go.Data) eventual2go.Data {
	return d.(serf.MemberEvent).Members[0]
}
