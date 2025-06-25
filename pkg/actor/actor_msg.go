package actor

// ActorMsg is the message format for all actor communication.
type ActorMsg struct {
	To        string
	From      string
	Type      string
	Data      interface{}
	ReplyChan chan ActorMsg `json:"-"`
}

// Factory functions for each known ActorMsg type.
// These functions document the arguments and centralize message creation.

// NewEvalMsg creates an ActorMsg for evaluating code.
func NewEvalMsg(to, from string, code string, replyChan chan ActorMsg) ActorMsg {
	return ActorMsg{
		To:        to,
		From:      from,
		Type:      "eval",
		Data:      code,
		ReplyChan: replyChan,
	}
}

// NewEvalResultMsg creates an ActorMsg for returning an eval result.
func NewEvalResultMsg(to, from string, result interface{}) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "eval_result",
		Data: result,
	}
}

// NewEvalErrorMsg creates an ActorMsg for returning an eval error.
func NewEvalErrorMsg(to, from string, errMsg string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "eval_error",
		Data: errMsg,
	}
}

// NewCreateChildMsg creates an ActorMsg to request creation of a child actor.
func NewCreateChildMsg(to, from, childType string, data interface{}, replyChan chan ActorMsg) ActorMsg {
	return ActorMsg{
		To:        to,
		From:      from,
		Type:      "create_child: " + childType,
		Data:      data,
		ReplyChan: replyChan,
	}
}

// NewChildCreatedMsg creates an ActorMsg to notify that a child was created.
func NewChildCreatedMsg(to, from, childName string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "child_created",
		Data: childName,
	}
}

// NewStopChildMsg creates an ActorMsg to request stopping a child actor.
func NewStopChildMsg(to, from, childName string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "stop_child",
		Data: childName,
	}
}

// NewForwardMessageMsg creates an ActorMsg to forward another message.
func NewForwardMessageMsg(to, from string, innerMsg ActorMsg) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "forward_message",
		Data: innerMsg,
	}
}

// NewConnectToPeerMsg creates an ActorMsg to connect to a peer.
func NewConnectToPeerMsg(to, from, peerURL string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "connect_to_peer",
		Data: peerURL,
	}
}

// NewReceiveResultMsg creates an ActorMsg for a successful receive function result.
func NewReceiveResultMsg(to, from string, result interface{}) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "receive_result",
		Data: result,
	}
}

// NewReceiveErrorMsg creates an ActorMsg for a failed receive function.
func NewReceiveErrorMsg(to, from string, errMsg string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "receive_error",
		Data: errMsg,
	}
}

// NewWebSocketRequestMsg creates an ActorMsg for a WebSocket request.
func NewWebSocketRequestMsg(to, from string, request interface{}) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "websocket_request",
		Data: request,
	}
}

// NewTestMessageMsg creates an ActorMsg for test messages.
func NewTestMessageMsg(to, from, data string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "test_message",
		Data: data,
	}
}

// NewPingMsg creates an ActorMsg for ping messages.
func NewPingMsg(to, from string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "ping",
	}
}

// NewPongMsg creates an ActorMsg for pong messages.
func NewPongMsg(to, from string) ActorMsg {
	return ActorMsg{
		To:   to,
		From: from,
		Type: "pong",
	}
}
