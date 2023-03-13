package gamer

import (
	"stress/network/pb"
	"time"

	"github.com/Gonewithmyself/gobot"
	"github.com/Gonewithmyself/gobot/pkg/btree"
	"github.com/Gonewithmyself/gobot/pkg/logger"

	"github.com/fish-tennis/gnet"
	"go.uber.org/zap"
)

type state int8

func (s state) String() string {
	if s < stateLoginOk {
		return "offline"
	}
	return "online"
}

const (
	stateIdle           state = iota // 离线
	stateLogining                    // 账号登录中
	stateLoginOk                     // 账号登录成功
	stateConnectedLogic              // 连接上logic
	stateEntrying                    // 角色登录中
	stateEntryOk                     // 角色登录成功
)

func (g *Gamer) getConnectionConfig() *gnet.ConnectionConfig {
	return &gnet.ConnectionConfig{
		SendPacketCacheCap: 16,
		SendBufferSize:     1024 * 10,
		RecvBufferSize:     1024 * 10,
		MaxPacketSize:      1024 * 10,
		RecvTimeout:        0,
		HeartBeatInterval:  5,
		WriteTimeout:       0,
	}
}

// 账号登录
func (g *Gamer) LoginAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	if g.status > stateIdle {
		return btree.SUCCESS
	}

	logger.Debug("LoginAction", "accountName", g.GetAccountName())

	g.conn = gnet.GetNetMgr().NewConnector(g.ctx, g.loginConf.Server, g.getConnectionConfig(),
		_clientCodec, _clientHandler, g)
	if g.conn == nil {
		logger.Error("Connect %v failed", g.loginConf.Server)
		return btree.ERROR
	}

	if !g.SendMsg(&pb.LoginReq{
		AccountName: g.GetAccountName(),
		Password:    g.GetAccountName()}) {
		logger.Error("LoginReq failed")
		g.conn.Close()
		g.conn.SetTag(nil)
		g.conn = nil
		return btree.ERROR
	}
	g.status = stateLogining
	return btree.SUCCESS

	//addr := strings.Split(g.authRsp.Addr, "://")[1]
	//client := network.NewClient(addr, g.authRsp.SdkUID, g)
	//client.GetConnection().SetTag(g)
	//logger.Debug("Connect")
	//if !client.Connect() {
	//	logger.Error("Connect failed")
	//	return btree.ERROR
	//}
	//
	//if !client.Send(&pb.LoginReq{
	//	AccountName: g.GetUid(),
	//	Password: "",
	//}) {
	//	logger.Error("LoginReq failed")
	//	return btree.ERROR
	//}
	//
	//go client.Run()
	//g.client = client

	//_, err := g.Server()
	//if err != nil {
	//	logger.Error("authErr", "err", err)
	//	return btree.ERROR
	//}

	//for _, zone := range rsp.Zones {
	//	if zone.GetId() == g.authData.Conf.Region {
	//		g.authRsp = rsp
	//		g.status = stateLoginOk
	//		g.ZoneName = *zone.Name
	//		return btree.SUCCESS
	//	}
	//}
	//g.Close()
	////logger.Error("zoneNotFound", "want", g.authData.Conf.Region, "got", rsp.Zones)
	//return btree.ERROR
}

func (g *Gamer) OnLoginRes(res *pb.LoginRes) {
	logger.Debug("onLoginRes", zap.Any("res", res))
	if res.Error == "NotReg" {
		// 自动注册账号
		// 这里是单纯的测试,账号和密码直接使用明文,实际项目需要做md5之类的处理
		g.SendMsg(&pb.AccountReg{
			AccountName: g.GetAccountName(),
			Password:    g.GetAccountName(),
		})
	} else if res.Error == "" {
		g.loginRes = res
		g.conn.SetTag(nil)
		g.conn.Close()
		g.conn = nil
		g.status = stateLoginOk
		g.accountId = res.AccountId
		g.accountName = res.AccountName
	} else {
		g.conn.SetTag(nil)
		g.conn.Close()
		g.conn = nil
		g.status = stateIdle
	}
}

func (g *Gamer) OnAccountRes(res *pb.AccountRes) {
	logger.Debug("onAccountRes", zap.Any("res", res))
	if res.Error == "" {
		g.SendMsg(&pb.LoginReq{
			AccountName: g.accountName,
			Password:    g.accountName,
		})
	}
}

func (g *Gamer) OnCoinRes(res *pb.CoinRes) {
}

func (g *Gamer) OnPlayerEntryGameRes(res *pb.PlayerEntryGameRes) {
	logger.Debug("OnPlayerEntryGameRes", zap.Any("res", res))
	if res.Error == "" {
		g.playerEntryGameRes = res
		g.status = stateEntryOk
		g.accountId = res.AccountId
		g.playerId = res.PlayerId
		g.region = res.RegionId
		g.playerName = res.PlayerName
		g.changeStatus("success")
		return
	}
	// 还没角色,则创建新角色
	if res.Error == "NoPlayer" {
		g.SendMsg(&pb.CreatePlayerReq{
			AccountId:    g.loginRes.GetAccountId(),
			LoginSession: g.loginRes.GetLoginSession(),
			RegionId:     1,
			Name:         g.accountName,
			Gender:       1,
		})
		return
	}
	// 登录遇到问题,服务器提示客户端稍后重试
	if res.Error == "TryLater" {
		// 延迟重试
		time.AfterFunc(time.Second, func() {
			g.conn.Send(gnet.PacketCommand(pb.CmdLogin_Cmd_PlayerEntryGameReq), &pb.PlayerEntryGameReq{
				AccountId:    g.loginRes.GetAccountId(),
				LoginSession: g.loginRes.GetLoginSession(),
				RegionId:     1,
			})
		})
	}
}

// 角色登录
func (g *Gamer) EntryAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	if g.status != stateLoginOk {
		return btree.SUCCESS
	}
	// 账号登录成功后,连接游戏服
	g.conn = gnet.GetNetMgr().NewConnector(g.ctx, g.loginRes.GetGameServer().GetClientListenAddr(), g.getConnectionConfig(),
		_clientCodec, _clientHandler, g)
	if g.conn == nil {
		logger.Error("%v connect game failed", g.GetAccountName())
		return btree.ERROR
	}
	g.SendMsg(&pb.PlayerEntryGameReq{
		AccountId:    g.loginRes.GetAccountId(),
		LoginSession: g.loginRes.GetLoginSession(),
		RegionId:     g.loginConf.Region,
	})

	g.status = stateEntrying
	return btree.SUCCESS
}

func (g *Gamer) HeartBeatAction(node *gobot.Worker, tick *btree.Tick) btree.Status {
	//g.SendMsg(&pb.GamerHeartC2S{})
	return btree.SUCCESS
}
