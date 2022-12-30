package gamer

import (
	"context"
	"fmt"
	"github.com/fish-tennis/gnet"
	"gobot/back"
	"reflect"
	"stress/network/pb"
	"stress/types"
)

// 一个玩家
type Gamer struct {
	*back.Gamer
	loginConf   *types.LoginConfig
	Seq         int32
	accountName string // 账号名
	accountId   int64
	playerId    int64
	name        string // 角色名
	lvl         int32
	region      int32 // 区服
	status      state
	ctx         context.Context
	conn        gnet.Connection

	loginRes           *pb.LoginRes           // 账号登录返回数据
	playerEntryGameRes *pb.PlayerEntryGameRes // 角色登录返回数据
}

type Role struct {
	Name     string
	Gid      int64
	Levle    int32
	ZoneName string
}

func NewGamer(ctx context.Context, conf *types.LoginConfig) *Gamer {
	account := fmt.Sprintf("%v", conf.Account)
	return &Gamer{
		loginConf:   conf,
		accountName: account,
		Gamer:       back.NewGamer(),
		ctx:         ctx,
	}
}

func NewGamerBySeq(ctx context.Context, seq int32, conf *types.LoginConfig) *Gamer {
	return &Gamer{
		loginConf:   conf,
		Seq:         seq,
		accountName: fmt.Sprintf("%v_%v", conf.Account, seq),
		Gamer:       back.NewGamer(),
		ctx:         ctx,
	}
}

// 玩家唯一标识
func (g *Gamer) GetAccountName() string {
	return g.accountName
}

func (g *Gamer) GetUid() string {
	return g.accountName
}

// 独立的玩家协程中处理网络消息
func (g *Gamer) ProcessMsg(data interface{}) {
	packet := data.(gnet.Packet)
	if protoPacket, ok := packet.(*gnet.ProtoPacket); ok {
		handlerMethod, ok2 := _clientHandler.methods[protoPacket.Command()]
		if ok2 {
			handlerMethod.Func.Call([]reflect.Value{reflect.ValueOf(g), reflect.ValueOf(protoPacket.Message())})
			return
		}
	}
	_clientHandler.DefaultConnectionHandler.OnRecvPacket(g.conn, packet)
	////TODO:统计消息
	//idx := strings.Index(msgName, "S2C")
	//if proto.MessageType(msgName[:idx]+"C2S") == nil {
	//	g.LogNtf(msgName, sc)
	//} else {
	//	g.LogRsp(msgName, sc)
	//}

	//msg := data.(gnet.Packet)
	//msgName := ""
	//defer func() {
	//	r := recover()
	//	if r != nil {
	//		g.LogError(msgName, fmt.Sprintf("%v", msg))
	//		logger.Error("procMsg", "name", msgName, "msg", msg)
	//	}
	//}()
	//if msg.Pkt == nil {
	//	// 第一个包 或者 新加了协议没编译
	//	logger.Debug("nilPkt", "cmd", msg.Hd.CmdAct())
	//	return
	//}
	//
	//msgName = reflect.TypeOf(msg.Pkt).Elem().Name()
	//// logger.Debug("recv", "msg", msgName)
	//router.Handle(msg, g)
}

func (g *Gamer) OnExit() {
	//g.client.Close()
	g.status = stateIdle
	//g.client = nil
	g.changeStatus("error")
}

func (g *Gamer) changeStatus(status string) {
	str := "离线"
	if g.IsOnline() {
		str = "在线"
	}
	g.ChangeStatus(g.name,
		status,
		fmt.Sprintf("id(%v) lvl(%v) region(%v) | %v",
			g.playerId, g.lvl, g.region, str))
}

func (g *Gamer) IsOnline() bool {
	return g.status >= stateEntryOk
}
