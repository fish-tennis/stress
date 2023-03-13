package network

import (
	"fmt"
	"sync"

	"github.com/rcrowley/go-metrics"
	"github.com/tealeg/xlsx"
)

// 各消息情况
var RecorderMap sync.Map // map[msgid]*Recorder

type Recorder struct {
	UpCounter   metrics.Counter // 请求数量
	DownCounter metrics.Counter // 回包数量
	RTTRecorder metrics.Timer   // 往返时延
}

func NewRecorder(msgId interface{}) *Recorder {
	str := fmt.Sprintf("%v", msgId)
	UpCounter := metrics.NewCounter()
	DownCounter := metrics.NewCounter()
	upval := metrics.GetOrRegister("Up_"+str, UpCounter)
	downVal := metrics.GetOrRegister("Down_"+str, DownCounter)

	tmpRtt := metrics.NewTimer()
	rtt := metrics.GetOrRegister("RTT_"+str, tmpRtt)
	return &Recorder{
		RTTRecorder: rtt.(metrics.Timer),
		UpCounter:   upval.(metrics.Counter),
		DownCounter: downVal.(metrics.Counter),
	}
}

type ReportItem struct {
	UpCount     int64  // 发包数量
	DownCount   int64  // 收包数量
	RspRate     string // 响应率
	Min         string // 最小时延 毫秒
	Max         string // 最大时延
	Avg         string // 平均时延
	Fifty       string // 时延50%分位
	SeventyFive string // 时延75%分位
	Ninety      string // 时延90%分位
}

// 输出报告
func Report() {
	data := Status()
	if len(data) == 0 {
		return
	}
	file := xlsx.NewFile()
	defer func() {
		file.Save("report.xlsx")
	}()
	sheet, _ := file.AddSheet("响应时间(毫秒)")

	headerFilds := []string{
		"消息名",
		"请求数量",
		"响应数量",
		"响应率",
		"最小响应",
		"最大响应",
		"平均响应",
		"50%分位",
		"75%分位",
		"90%分位",
	}
	row := sheet.AddRow() // 表头
	for _, fd := range headerFilds {
		row.AddCell().Value = fd
	}

	// 数据
	for msgName, info := range Status() {
		if info.UpCount == 0 {
			continue
		}

		row := sheet.AddRow()
		row.AddCell().Value = msgName
		row.AddCell().SetValue(info.UpCount)
		row.AddCell().SetValue(info.DownCount)
		row.AddCell().SetValue(info.RspRate)
		row.AddCell().SetValue(info.Min)
		row.AddCell().SetValue(info.Max)
		row.AddCell().SetValue(info.Avg)
		row.AddCell().SetValue(info.Fifty)
		row.AddCell().SetValue(info.SeventyFive)
		row.AddCell().SetValue(info.Ninety)
	}
}

// 输出当前的统计信息
func Status() map[string]*ReportItem {
	var mp = make(map[string]*ReportItem)
	RecorderMap.Range(func(key, value interface{}) bool {
		rec := value.(*Recorder)
		item := &ReportItem{
			UpCount:     rec.UpCounter.Count(),
			DownCount:   rec.DownCounter.Count(),
			Min:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Min())/1e6),
			Max:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Max())/1e6),
			Avg:         fmt.Sprintf("%.2f", float64(rec.RTTRecorder.Mean())/1e6),
			Fifty:       fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.5)/1e6),
			SeventyFive: fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.75)/1e6),
			Ninety:      fmt.Sprintf("%.2f", rec.RTTRecorder.Percentile(0.90)/1e6),
		}
		if rec.UpCounter.Count() == 0 {
			item.RspRate = "0"
		} else {
			dt := float64(rec.DownCounter.Count()) / float64(rec.UpCounter.Count())
			item.RspRate = fmt.Sprintf("%.2f%%", dt*100)
		}
		mp[GetMessageNameById(key.(uint16))] = item
		return true
	})
	return mp
}

func GetRecorder(msgId uint16) *Recorder {
	val, ok := RecorderMap.Load(msgId)
	if ok {
		return val.(*Recorder)
	}

	rec := NewRecorder(msgId)
	val, _ = RecorderMap.LoadOrStore(msgId, rec)
	return val.(*Recorder)
}
