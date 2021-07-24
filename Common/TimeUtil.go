package Common

import (
	"database/sql"
	"time"

	"github.com/gookit/slog"
)

// utc: Poco DateTime compitable
// Poco::utc - (int64(0x01b21dd2) << 32) - 0x13814000 = time.UnixNano()/100

func SqlString_to_utc(v sql.NullString) int64 {
	if !v.Valid {
		return 0
	}

	return String_to_utc(v.String)
}

func String_to_utc(str string) int64 {

	if len(str) == 0 || str == "0000-00-00 00:00:00" {
		return 0
	}

	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		slog.Error(err)
		return 0
	}

	return t.UnixNano()/100 + (int64(0x01b21dd2) << 32) + 0x13814000
}

func String_to_DateTime(str string) time.Time {

	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		slog.Error(err)
		return time.Unix(0, 0)
	}

	return t
}

func Utc_to_DateTime(utc int64) time.Time {

	nano := (utc - (int64(0x01b21dd2) << 32) - 0x13814000) * 100
	return time.Unix(0, nano)
}

func Utc_to_String(utc int64) string {

	t := Utc_to_DateTime(utc)
	return t.Format("2006-01-02 15:04:05")
}

func DateTime_to_utc(t time.Time) int64 {

	return t.UnixNano()/100 + (int64(0x01b21dd2) << 32) + 0x13814000
}

func DateTime_to_String(t time.Time) string {

	return t.Format("2006-01-02 15:04:05")
}

//----------------------------------------------------------------------------
type SimpleTimer struct {
	iTimeLength int64
	iTimeCur    int64
	bEnabled    bool
	bRepeat     bool
}

//时间单位毫秒
func (t *SimpleTimer) SetTimer(len int64, repeat bool, triggerAtOnce bool) {
	t.bEnabled = true
	t.bRepeat = repeat
	t.iTimeLength = len
	t.iTimeCur = 0
	if triggerAtOnce {
		t.iTimeCur = t.iTimeLength
	}
}

//重置当前时间, 时间周期不变
func (t *SimpleTimer) ResetTimer() {
	t.iTimeCur = 0
}

//传入间隔时间，毫秒. 返回Timer是否触发.
func (t *SimpleTimer) OnTimer(iDelta int64) bool {
	if !t.bEnabled {
		return false
	}

	t.iTimeCur = t.iTimeCur + iDelta
	if t.iTimeCur >= t.iTimeLength {
		if t.bRepeat {
			t.iTimeCur = 0
		} else {
			t.bEnabled = false
		}
		return true
	}

	return false
}

//Timer是否在工作.
func (t *SimpleTimer) IsTimerSet() bool {
	return t.bEnabled
}

//禁止该Timer.
func (t *SimpleTimer) CancelTimer() {
	t.bEnabled = false
	t.iTimeCur = 0
}

//把周期延长一定时间, 毫秒.
func (t *SimpleTimer) Postpond(dt int64) {
	t.iTimeLength = t.iTimeLength + dt
}

//前进一定时间, 毫秒.
func (t *SimpleTimer) Advance(dt int64) {
	t.iTimeCur = t.iTimeCur + dt
}

//获取剩余时间, 毫秒.
func (t *SimpleTimer) GetLeftTime() int64 {
	if !t.bEnabled {
		return 0
	}
	return t.iTimeLength - t.iTimeCur
}

//获取时间周期, 毫秒.
func (t *SimpleTimer) GetCycleTime() int64 {
	return t.iTimeLength
}
