package network

import (
	"fmt"
	"go.uber.org/zap"
	"gobot/pkg/logger"
	"gobot/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"reflect"
	"strings"
)

var AppPbInfo AppPb

// 预处理pb协议
type AppPb struct {
	CsList []string               // 所有cs消息名
	CsType2Cmd map[reflect.Type]uint16
	CsCmd2Type map[uint16]reflect.Type
	ScCmd2Type map[uint16]reflect.Type
	Name2Type map[string]protoreflect.MessageType
	Special    map[string]*util.SpecialInfo
}

func GetName(cmd uint16) string {
	tp, ok := AppPbInfo.ScCmd2Type[cmd]
	if !ok {
		return ""
	}
	return tp.Name()
}

func (info *AppPb) ListMsg() []string {
	return info.CsList
}

func (info *AppPb) GetMsgDefault(name string) interface{} {
	logger.Debug("GetMsgDefault",zap.String("name",name))
	if typ,ok := info.Name2Type[name]; ok {
		newMessage := typ.New()
		if newMessage == nil {
			logger.Error("GetMsgDefault new err", zap.String("name",name))
			return nil
		}
		protoMessage,ok2 := newMessage.Interface().(proto.Message)
		if !ok2 {
			logger.Error("GetMsgDefault convert err", zap.String("name",name))
			return nil
		}
		logger.Debug("GetMsgDefault json",zap.String("json",
			protojson.MarshalOptions{EmitUnpopulated: true,UseProtoNames:true}.Format(protoMessage)))
		return protojson.MarshalOptions{EmitUnpopulated: true,UseProtoNames:true}.Format(protoMessage)
	}
	return nil
}

func (info *AppPb) GetCsMsgByJSON(name string, js string) proto.Message {
	if typ,ok := info.Name2Type[name]; ok {
		newMessage := typ.New()
		if newMessage == nil {
			logger.Error("GetMsgDefault new err", zap.String("name",name))
			return nil
		}
		protoMessage,ok2 := newMessage.Interface().(proto.Message)
		if !ok2 {
			logger.Error("GetMsgDefault convert err", zap.String("name",name))
			return nil
		}
		err := protojson.UnmarshalOptions{}.Unmarshal([]byte(js), protoMessage)
		if err != nil {
			logger.Error("Unmarshal err", zap.Error(err), zap.String("name",name), zap.String("js", js))
			return nil
		}
		logger.Debug(fmt.Sprintf("%v", protoMessage))
		return protoMessage
	}
	return nil
}

func (info *AppPb) Init() {
	info.CsCmd2Type = make(map[uint16]reflect.Type)
	info.CsType2Cmd = make(map[reflect.Type]uint16)
	info.ScCmd2Type = make(map[uint16]reflect.Type)
	info.Name2Type = make(map[string]protoreflect.MessageType)
	info.Special = make(map[string]*util.SpecialInfo)

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
			elem := messageType.New()
			typ := reflect.TypeOf(elem).Elem()
			info.CsType2Cmd[typ] = uint16(messageId)
			if strings.HasSuffix(messageName, "Req") {
				info.CsCmd2Type[uint16(messageId)] = typ
			} else {
				info.ScCmd2Type[uint16(messageId)] = typ
			}
			info.CsList = append(info.CsList, messageName)
			info.Name2Type[messageName] = messageType
			//dft := util.JsonDefault(typ)
			//info.Special[messageName] = dft
			//logger.Debug("messageDescriptor", zap.Int32("messageId", messageId), zap.String("name", messageName))
		}
		return true
	})
}
