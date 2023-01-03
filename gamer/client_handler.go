package gamer

import (
	"fmt"
	"gobot/pkg/logger"
	"reflect"
	"stress/network/pb"
	"strings"
	"time"

	"github.com/fish-tennis/gnet"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var _clientHandler *ClientHandler
var _clientCodec *gnet.ProtoCodec

type ClientHandler struct {
	gnet.DefaultConnectionHandler
	methods map[gnet.PacketCommand]reflect.Method
}

func NewClientHandler(protoCodec *gnet.ProtoCodec) *ClientHandler {
	handler := &ClientHandler{
		DefaultConnectionHandler: *gnet.NewDefaultConnectionHandler(protoCodec),
		methods:                  make(map[gnet.PacketCommand]reflect.Method),
	}
	handler.RegisterHeartBeat(gnet.PacketCommand(pb.CmdInner_Cmd_HeartBeatReq), func() proto.Message {
		return &pb.HeartBeatReq{
			Timestamp: uint64(time.Now().UnixNano() / int64(time.Millisecond)),
		}
	})
	handler.Register(gnet.PacketCommand(pb.CmdInner_Cmd_HeartBeatRes), func(connection gnet.Connection, packet *gnet.ProtoPacket) {
	}, new(pb.HeartBeatRes))
	handler.SetUnRegisterHandler(func(connection gnet.Connection, packet *gnet.ProtoPacket) {
		logger.Debug(fmt.Sprintf("un register %v", string(packet.Message().ProtoReflect().Descriptor().Name())))
	})
	return handler
}

func (this *ClientHandler) OnRecvPacket(connection gnet.Connection, packet gnet.Packet) {
	if connection.GetTag() != nil {
		g := connection.GetTag().(*Gamer)
		g.OnRecvPacket(packet)
		//if protoPacket, ok := packet.(*gnet.ProtoPacket); ok {
		//	handlerMethod, ok2 := this.methods[protoPacket.Command()]
		//	if ok2 {
		//		handlerMethod.Func.Call([]reflect.Value{reflect.ValueOf(g), reflect.ValueOf(protoPacket.Message())})
		//		return
		//	}
		//}
		//this.DefaultConnectionHandler.OnRecvPacket(connection, packet)
		// TODO:统计消息
		//idx := strings.Index(msgName, "S2C")
		//if proto.MessageType(msgName[:idx]+"C2S") == nil {
		//	g.LogNtf(msgName, sc)
		//} else {
		//	g.LogRsp(msgName, sc)
		//}
	}
}

func InitClientHandler() {
	_clientCodec = gnet.NewProtoCodec(nil)
	_clientHandler = NewClientHandler(_clientCodec)
	_clientHandler.autoRegister()
	_clientHandler.SetOnDisconnectedFunc(func(connection gnet.Connection) {
		if connection.GetTag() == nil {
			return
		}
		g := connection.GetTag().(*Gamer)
		if g != nil && g.conn == connection {
			connection.SetTag(nil)
			g.conn = nil
		}
		logger.Debug(fmt.Sprintf("client disconnect %v", g.GetAccountName()))
	})
}

// 通过反射自动注册消息回调
func (this *ClientHandler) autoRegister() {
	typ := reflect.TypeOf(&Gamer{})
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		// func (this *Gamer) OnMessageName(res *pb.MessageName)
		if method.Type.NumIn() != 2 || !strings.HasPrefix(method.Name, "On") {
			continue
		}
		methonArg1 := method.Type.In(1)
		if !strings.HasPrefix(methonArg1.String(), "*pb.") {
			continue
		}
		// 消息名,如: LoginRes
		messageName := methonArg1.String()[strings.LastIndex(methonArg1.String(), ".")+1:]
		// 函数名必须是onLoginRes
		if method.Name != fmt.Sprintf("On%v", messageName) {
			logger.Debug(fmt.Sprintf("methodName not match:%v", method.Name))
			continue
		}
		// Cmd_LoginRes
		enumValueName := fmt.Sprintf("Cmd_%v", messageName)
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
			continue
		}
		cmd := gnet.PacketCommand(messageId)
		// 注册消息回调到组件上
		this.methods[cmd] = method
		// 注册消息的构造函数
		this.DefaultConnectionHandler.Register(cmd, nil, reflect.New(methonArg1.Elem()).Interface().(proto.Message))
		logger.Debug(fmt.Sprintf("register %v %v", messageId, method.Name))
	}
}
