package network

import (
	"fmt"
	"strings"

	"github.com/Gonewithmyself/gobot/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

var AppPbInfo AppPb

// 预处理pb协议
type AppPb struct {
	ReqMessageNameList []string // 所有cs消息名
	NameMessageTypeMap map[string]protoreflect.MessageType
	CmdNameMap         map[uint16]string
	ResId2Req          map[uint16]uint16 //
}

func GetMessageNameById(messageId uint16) string {
	name, ok := AppPbInfo.CmdNameMap[messageId]
	if !ok {
		return ""
	}
	return name
}

func (info *AppPb) ListMsg() []string {
	return info.ReqMessageNameList
}

func (info *AppPb) HasReqMessage(resMessageName string) bool {
	idx := strings.Index(resMessageName, "Res")
	if idx < 0 {
		return false
	}
	if _, ok := info.NameMessageTypeMap[resMessageName[:idx]+"Req"]; ok {
		return true
	}
	return false
}

func (info *AppPb) GetMsgDefault(name string) interface{} {
	logger.Debug("GetMsgDefault", zap.String("name", name))
	if typ, ok := info.NameMessageTypeMap[name]; ok {
		newMessage := typ.New()
		if newMessage == nil {
			logger.Error("GetMsgDefault new err", zap.String("name", name))
			return nil
		}
		protoMessage, ok2 := newMessage.Interface().(proto.Message)
		if !ok2 {
			logger.Error("GetMsgDefault convert err", zap.String("name", name))
			return nil
		}
		logger.Debug("GetMsgDefault json", zap.String("json",
			protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}.Format(protoMessage)))
		return protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}.Format(protoMessage)
	}
	return nil
}

func (info *AppPb) GetCsMsgByJSON(name string, js string) proto.Message {
	if typ, ok := info.NameMessageTypeMap[name]; ok {
		newMessage := typ.New()
		if newMessage == nil {
			logger.Error("GetMsgDefault new err", zap.String("name", name))
			return nil
		}
		protoMessage, ok2 := newMessage.Interface().(proto.Message)
		if !ok2 {
			logger.Error("GetMsgDefault convert err", zap.String("name", name))
			return nil
		}
		err := protojson.UnmarshalOptions{}.Unmarshal([]byte(js), protoMessage)
		if err != nil {
			logger.Error("Unmarshal err", zap.Error(err), zap.String("name", name), zap.String("js", js))
			return nil
		}
		logger.Debug(fmt.Sprintf("%v", protoMessage))
		return protoMessage
	}
	return nil
}

func (info *AppPb) Init() {
	info.ResId2Req = make(map[uint16]uint16)
	info.NameMessageTypeMap = make(map[string]protoreflect.MessageType)
	info.CmdNameMap = make(map[uint16]string)

	name2IdMap := make(map[string]uint16)
	protoregistry.GlobalFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		for i := 0; i < fileDescriptor.Messages().Len(); i++ {
			messageDescriptor := fileDescriptor.Messages().Get(i)
			messageName := string(messageDescriptor.Name())
			if !strings.HasSuffix(messageName, "Req") && !strings.HasSuffix(messageName, "Res") {
				continue
			}
			//logger.Debug("messageDescriptor", zap.String("name", messageName))
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
			messageType, err := protoregistry.GlobalTypes.FindMessageByName(messageDescriptor.FullName())
			if err != nil {
				continue
			}

			info.ReqMessageNameList = append(info.ReqMessageNameList, messageName)
			info.NameMessageTypeMap[messageName] = messageType
			info.CmdNameMap[uint16(messageId)] = messageName
			name2IdMap[messageName] = uint16(messageId)
			//logger.Debug("messageDescriptor", zap.Int32("messageId", messageId), zap.String("name", messageName))
		}
		return true
	})

	for msgName, msgId := range name2IdMap {
		if strings.HasSuffix(msgName, "Res") {
			reqName := strings.TrimSuffix(msgName, "Res") + "Req"
			info.ResId2Req[msgId] = name2IdMap[reqName]
		}
	}
}
