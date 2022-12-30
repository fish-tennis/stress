package gamer

import (
	"fmt"
	"github.com/fish-tennis/gnet"
	"gobot/pkg/ratelimit"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"reflect"
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
	//	name := method.Name
	//	switch {
	//	case strings.HasSuffix(name, "Res"):
	//		// 消息回调
	//		msgtp := proto.MessageType(name)
	//		if msgtp == nil {
	//			panic(fmt.Sprintf("msg(%v)NotFound", name))
	//		}
	//		router[msgtp.Elem()] = fn.(func(*Gamer, interface{}))
	//
	//	case strings.HasPrefix(name, "Err"):
	//		// 错误码回调
	//		errRouter[name] = fn.(func(*Gamer, string, string))
	//	}
	//}
}
