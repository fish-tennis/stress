package network

import (
	"gobot/pkg/metric"
	"time"
)

const (
	DirectionUp   = 1
	DirectionDown = 2
)

type Metric struct {
	Gamer  string // 玩家uid
	CmdAct uint16 // 命令
	Direct uint16 // 上行还是下行
	Bytes  int64  // 消息大小
	Ts     int64  // 时间戳
}

type MetricHandler struct {
	metric.Reporter
}

func (h *MetricHandler) TransMsgId(id interface{}) string {
	cmdact := id.(uint16)
	return GetName(cmdact)
}

func (h *MetricHandler) GetReportFile() string {
	return "report.xlsx"
}

type rttmap map[uint16]int64          // cmd请求发送时间戳
var gamerMetric = map[string]rttmap{} // 玩家指标

// 处理指标
// 根据gamerid 映射固定的协程中调用
func (h *MetricHandler) ProcessMetric(v metric.IMetric, rec *metric.Recorder) {
	mtr := v.(*Metric)
	gamerData, ok := gamerMetric[mtr.Gamer]
	if !ok {
		gamerData = rttmap{}
		gamerMetric[mtr.Gamer] = gamerData
	}

	if mtr.Direct == DirectionUp {
		// 发起请求
		gamerData[mtr.CmdAct] = mtr.Ts
		rec.UpCounter.Inc(1)
		return
	}

	// 收到回包
	rec.DownCounter.Inc(1)
	reqTs := gamerData[mtr.CmdAct]
	if reqTs == 0 {
		return
	}
	rtt := mtr.Ts - reqTs
	rec.RTTRecorder.Update(time.Nanosecond * time.Duration(rtt))
	gamerData[mtr.CmdAct] = 0
}

func (mtr *Metric) GetGamer() string {
	return mtr.Gamer
}

func (mtr *Metric) GetMsgId() interface{} {
	return mtr.CmdAct
}
