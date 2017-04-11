package models

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// MinimumTick 毫秒精度
	minimumTick = time.Millisecond
	// second 秒精度
	second = int64(time.Second / minimumTick)
	// nanosPerTick 纳秒精度
	nanosPerTick = int64(minimumTick / time.Nanosecond)

	// Earliest 最小时间
	Earliest = Time(math.MinInt64)
	// Latest 最大时间
	Latest = Time(math.MaxInt64)
)

// Time 自定义时间类型
type Time int64

// Interval 自定义间隔类型
type Interval struct {
	Start, End Time
}

// Now 获取当前时间
func Now() Time {
	return TimeFromUnixNano(time.Now().UnixNano())
}

// TimeFromUnix 转换系统时间到自定义时间
func TimeFromUnix(t int64) Time {
	return Time(t * second)
}

// TimeFromUnixNano 系统纳秒到自定义时间
func TimeFromUnixNano(t int64) Time {
	return Time(t / nanosPerTick)
}

// Equal 比较
func (t Time) Equal(o Time) bool {
	return t == o
}

// Before 比较
func (t Time) Before(o Time) bool {
	return t < o
}

// After 比较
func (t Time) After(o Time) bool {
	return t > o
}

// Add 计算时间漂移
func (t Time) Add(d time.Duration) Time {
	return t + Time(d/minimumTick)
}

// Sub 计算时间差
func (t Time) Sub(o Time) time.Duration {
	return time.Duration(t-o) * minimumTick
}

// Time 获取系统时间
func (t Time) Time() time.Time {
	return time.Unix(int64(t)/second, (int64(t)%second)*nanosPerTick)
}

// Unix 系统时间戳（秒）
func (t Time) Unix() int64 {
	return int64(t) / second
}

// UnixNano 系统时间戳（纳秒）
func (t Time) UnixNano() int64 {
	return int64(t) * nanosPerTick
}

// The number of digits after the dot.
var dotPrecision = int(math.Log10(float64(second)))

// String 实现Stringer
func (t Time) String() string {
	return strconv.FormatFloat(float64(t)/float64(second), 'f', -1, 64)
}

// MarshalJSON json编码实现
func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalJSON json解码实现
func (t *Time) UnmarshalJSON(b []byte) error {
	p := strings.Split(string(b), ".")
	switch len(p) {
	case 1:
		v, err := strconv.ParseInt(string(p[0]), 10, 64)
		if err != nil {
			return err
		}
		*t = Time(v * second)

	case 2:
		// 如果是浮点数需要计算小数点位数
		v, err := strconv.ParseInt(string(p[0]), 10, 64)
		if err != nil {
			return err
		}
		// 整数秒
		v *= second

		prec := dotPrecision - len(p[1])
		if prec < 0 {
			p[1] = p[1][:dotPrecision]
		} else if prec > 0 {
			p[1] = p[1] + strings.Repeat("0", prec)
		}
		va, err := strconv.ParseInt(p[1], 10, 32)
		if err != nil {
			return err
		}

		*t = Time(v + va)

	default:
		return fmt.Errorf("invalid time %q", string(b))
	}
	return nil
}

// Duration 自定义Duration，支持类是cron的格式解析
type Duration time.Duration

var durationRE = regexp.MustCompile("^([0-9]+)(y|w|d|h|m|s|ms)$")

// ParseDuration 从字符串解析
func ParseDuration(durationStr string) (Duration, error) {
	matches := durationRE.FindStringSubmatch(durationStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("not a valid duration string: %q", durationStr)
	}
	var (
		n, _ = strconv.Atoi(matches[1])
		dur  = time.Duration(n) * time.Millisecond
	)
	switch unit := matches[2]; unit {
	case "y":
		dur *= 1000 * 60 * 60 * 24 * 365
	case "w":
		dur *= 1000 * 60 * 60 * 24 * 7
	case "d":
		dur *= 1000 * 60 * 60 * 24
	case "h":
		dur *= 1000 * 60 * 60
	case "m":
		dur *= 1000 * 60
	case "s":
		dur *= 1000
	case "ms":
		// Value already correct
	default:
		return 0, fmt.Errorf("invalid time unit in duration string: %q", unit)
	}
	return Duration(dur), nil
}

// String 实现stringer
func (d Duration) String() string {
	var (
		ms   = int64(time.Duration(d) / time.Millisecond)
		unit = "ms"
	)
	factors := map[string]int64{
		"y":  1000 * 60 * 60 * 24 * 365,
		"w":  1000 * 60 * 60 * 24 * 7,
		"d":  1000 * 60 * 60 * 24,
		"h":  1000 * 60 * 60,
		"m":  1000 * 60,
		"s":  1000,
		"ms": 1,
	}

	switch int64(0) {
	case ms % factors["y"]:
		unit = "y"
	case ms % factors["w"]:
		unit = "w"
	case ms % factors["d"]:
		unit = "d"
	case ms % factors["h"]:
		unit = "h"
	case ms % factors["m"]:
		unit = "m"
	case ms % factors["s"]:
		unit = "s"
	}
	return fmt.Sprintf("%v%v", ms/factors[unit], unit)
}

// MarshalJSON 实现json编码.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalJSON 实现json解码.
func (d *Duration) UnmarshalJSON(b []byte) error {
	dur, err := ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = dur
	return nil
}
