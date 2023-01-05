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

type (
	Handler    map[reflect.Type]func(*Gamer, interface{}) // 消息回调
	ErrHandler map[string]func(*Gamer, string, string)    // 错误码回调
)

func (r Handler) Handle(msg gnet.Packet, g *Gamer) {
	//sc := msg.Pkt
	//tp := reflect.TypeOf(sc).Elem()
	//msgName := tp.Name()
	//if code := msg.Hd.Error; code != 0 {
	//	//// 错误码处理
	//	//codeName := pb.ErrorCode_name[int32(code)]
	//	//if h, ok := errRouter[codeName]; ok {
	//	//	h(g, codeName, msgName)
	//	//}
	//	return
	//}
	//
	//h, ok := r[tp]
	//if ok {
	//	h(g, sc)
	//}
	//
	//idx := strings.Index(msgName, "S2C")
	//if proto.MessageType(msgName[:idx]+"C2S") == nil {
	//	g.LogNtf(msgName, sc)
	//} else {
	//	g.LogRsp(msgName, sc)
	//}
}

var (
	router    = Handler{}
	errRouter = ErrHandler{}
)

func init() {
	autoRegisterHandler()
}

func autoRegisterHandler() {
	//tp := reflect.TypeOf(&Gamer{})
	//for i := 0; i < tp.NumMethod(); i++ {
	//	method := tp.Method(i)
	//	fn := method.Func.Interface()
	//	playerName := method.Name
	//	switch {
	//	case strings.HasSuffix(playerName, "Res"):
	//		// 消息回调
	//		msgtp := proto.MessageType(playerName)
	//		if msgtp == nil {
	//			panic(fmt.Sprintf("msg(%v)NotFound", playerName))
	//		}
	//		router[msgtp.Elem()] = fn.(func(*Gamer, interface{}))
	//
	//	case strings.HasPrefix(playerName, "Err"):
	//		// 错误码回调
	//		errRouter[playerName] = fn.(func(*Gamer, string, string))
	//	}
	//}
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
