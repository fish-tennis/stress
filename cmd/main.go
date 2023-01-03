package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"stress/gamer"
	"stress/network"
	"stress/types"
	"time"

	"github.com/Gonewithmyself/gobot"
	"github.com/Gonewithmyself/gobot/back"
	"github.com/Gonewithmyself/gobot/pkg/logger"
	"github.com/Gonewithmyself/gobot/pkg/metric"
	"github.com/Gonewithmyself/gobot/pkg/util"

	"github.com/fish-tennis/gnet"
	"go.uber.org/zap/zapcore"
)

const (
	staticDir = "./static"
	confDir   = "./conf"
	treeFile  = "robot.b3"
)

func main() {
	// 命令行参数解析
	util.ParseCmdArgs(&types.Args)

	// 读取配置
	ReadConfig()

	// 设置logger
	SetupLogger()

	// 注册行为树叶子节点实现
	gobot.RegisterAction(&gamer.Gamer{})

	app := &App{}

	// linux下压测时 通过命令行参数传递压测时间
	ctx := context.Background()
	app.ctx = ctx
	if types.Args.Timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, time.Second*time.Duration(types.Args.Timeout))
	}

	network.AppPbInfo.Init()
	gamer.InitClientHandler()

	metric.Start(&network.MetricHandler{})
	defer metric.Stop()
	// 开始运行
	gobot.RunApp(
		ctx,
		app,
		gobot.WithStaticDir(staticDir),
		gobot.WithTreeFile(filepath.Join(confDir, treeFile)),
		gobot.WithStopWaitSec(int64(types.Args.StopWait)),
	)

}

// 读取配置app.json
func ReadConfig() {
	data, err := ioutil.ReadFile(filepath.Join(confDir, "app.json"))
	if err != nil {
		panic(err)
	}

	var tmp types.AppConfig
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		panic(err)
	}
	types.AppConf = &tmp

}

func SetupLogger() {
	lv, er := zapcore.ParseLevel(types.AppConf.LogLevel)
	if er != nil {
		panic(er)
	}
	logger.Init(lv, "", "app")

	// gnet.SetLogger(logger.GetLogger(), int8(lv))
	gnet.SetLogLevel(int8(lv))
}

type App struct {
	gobot.App
	ctx context.Context
}

func (app *App) OnExit(reason string) {
	logger.Info("OnExit", "reason", reason)
}

// 根据配置构造玩家
func (app *App) CreateGamer(confJson string, seq int32) (back.IGamer, error) {
	loginConf, err := getLoginConf(confJson)
	if err != nil {
		return nil, err
	}
	if seq < 0 {
		// 运行单个机器人
		return gamer.NewGamer(app.ctx, loginConf), nil
	}
	return gamer.NewGamerBySeq(app.ctx, seq, loginConf), nil
}

// 开始压测
func (app *App) StressStart(start, count int32, treeID, confJs string) {
	if app.Os == gobot.OsWindows {
		// windows下 由UI构造压测参数
		app.BaseApp.StressStart(start, count, treeID, confJs)
		return
	}

	// linux下 由命令行构造参数
	args := types.Args
	conf := types.LoginConfig{
		Server:        args.Server,
		Region:        args.Region,
		Device:        "xxx",
		ClientVersion: "0.0.0.0.0.0",
	}
	js, _ := json.Marshal(conf)
	app.BaseApp.StressStart(args.Start, args.Count, args.Tree, string(js))
}

// 打印压测状态
func (app *App) PrintStressStatus() {
	if len(app.Gamers) == 0 {
		// app.UI.Print("压测0个机器人", nil)
		return
	}

	var status struct {
		Total      int32 `json:"总数量"`
		Online     int32 `json:"在线人数"`
		MetricData interface{}
	}

	for _, v := range app.Gamers {
		status.Total += 1
		gamer := v.(*gamer.Gamer)
		if gamer.IsOnline() {
			status.Online += 1
		}
	}

	status.MetricData = metric.Status()
	app.UI.Print("压测状态", status)
	app.UI.UIChangeStatus(0, "压测中", "success",
		fmt.Sprintf("online(%v) total(%v)", status.Online, status.Total))
}

/*
	给ui的接口
*/
// 热更配置
func (app *App) JsReloadConfig() (ret interface{}, err error) {
	logger.Debug("onReload")
	ReadConfig()
	return
}

// 区服列表
func (app *App) JsFetchZones(confJs string) (ret interface{}, err error) {
	//conf, err := getLoginConf(confJs)
	//if err != nil {
	//	return nil, err
	//}
	//account := fmt.Sprintf("%v", conf.Account)
	//auther := network.NewAuther(&types.AuthData{
	//	SdkAccount: account,
	//	Conf:       conf,
	//})
	//sdkdata, err := auther.SdkAuth(nil)
	//if err != nil {
	//	return nil, err
	//}
	//rsp, err := auther.GameAuth(sdkdata)
	//if err != nil {
	//	return nil, err
	//}
	//
	//ret = rsp.(*network.Authrsp).Zones
	return
}

// 拉取所有消息列表
func (app *App) JsGetMsgList() (ret interface{}, err error) {
	ret = network.AppPbInfo.ListMsg()
	return
}

// 获取消息详情
func (app *App) JsGetMsgDetail(msg string) (ret interface{}, err error) {
	ret = network.AppPbInfo.GetMsgDefault(msg)
	return
}

func (app *App) JsSendReq(name, data string) (ret interface{}, err error) {
	if len(app.Gamers) == 0 {
		err = fmt.Errorf("notLogin")
		return
	}

	msg := network.AppPbInfo.GetCsMsgByJSON(name, data)
	for _, v := range app.Gamers {
		g := v.(*gamer.Gamer)
		g.SendMsg(msg)
	}
	return
}

func getLoginConf(confJs string) (conf *types.LoginConfig, err error) {
	conf = &types.LoginConfig{}
	json.Unmarshal([]byte(confJs), conf)
	return
}
