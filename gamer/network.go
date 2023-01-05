package gamer

import (
	"fmt"
	"reflect"

	"github.com/Gonewithmyself/gobot/pkg/ratelimit"
	"github.com/fish-tennis/gnet"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// 收到网络层包 转发到玩家协程
func (g *Gamer) OnRecvPacket(packet gnet.Packet) {
	g.MsgCh <- packet
}

func (g *Gamer) SendMsg(msg proto.Message) bool {
	if g.conn == nil {
		return false
	}

	tp := reflect.TypeOf(msg).Elem()
	if !ratelimit.Consume(tp.Name()) {
		return false
	}

	enumValueName := fmt.Sprintf("Cmd_%v", tp.Name())
	var messageId int32
	protoregistry.GlobalTypes.RangeEnums(func(enumType protoreflect.EnumType) bool {
		// gserver.Login.CmdLogin
		enumValueDescriptor := enumType.Descriptor().Values().ByName(protoreflect.Name(enumValueName))
		if enumValueDescriptor != nil {
			messageId = int32(enumValueDescriptor.Number())
			return false
		}
		return true
	})
	if messageId == 0 {
		return false
	}

	if g.conn.Send(gnet.PacketCommand(messageId), msg) {
		g.LogReq(tp.Name(), msg)
		return true
	}

	return false
}
