package gamer

import (
	"fmt"
	"reflect"
	"stress/network"
	"time"

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
		// network.GetRecorder()
		g.LogReq(uint16(messageId), tp.Name(), msg)
		return true
	}

	return false
}

// 记录请求消息
func (g *Gamer) LogReq(msgId uint16, msgName string, msg interface{}) {
	g.Gamer.LogReq(msgName, msg) // 显示到日志区

	g.rttmap[msgId] = time.Now().UnixNano() // 记录请求发送时间
	rec := network.GetRecorder(msgId)
	rec.UpCounter.Inc(1)
}

// 记录响应消息
func (g *Gamer) LogRes(msgId uint16, msgName string, msg interface{}) {
	now := time.Now().UnixNano()
	g.Gamer.LogRsp(msgName, msg)

	reqMsgId, ok := network.AppPbInfo.ResId2Req[msgId]
	if !ok {
		return
	}

	rec := network.GetRecorder(reqMsgId)
	rec.DownCounter.Inc(1)
	reqTs := g.rttmap[reqMsgId]
	if reqTs == 0 {
		return
	}

	rtt := now - reqTs
	rec.RTTRecorder.Update(time.Nanosecond * time.Duration(rtt))
	g.rttmap[reqMsgId] = 0
}
