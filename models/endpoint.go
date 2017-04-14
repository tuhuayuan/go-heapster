package models

import (
	"encoding/json"
	"fmt"
	"net"
)

// Endpoint 服务器地址
type Endpoint string

// MarshalJSON json编码
func (ep Endpoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ep))
}

// UnmarshalJSON json解码
func (ep *Endpoint) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	rawEp := Endpoint(str)
	if err := rawEp.Validate(); err != nil {
		return err
	}
	*ep = rawEp
	return nil
}

// Equal 判断相等
func (ep Endpoint) Equal(other Endpoint) bool {
	return string(ep) == string(other)
}

// Validate 验证, 仅仅支持IPV4
func (ep Endpoint) Validate() error {
	ip := net.ParseIP(string(ep))
	if ip == nil && !ep.IsCIDRAddr() {
		return fmt.Errorf("error ip address")
	}
	return nil
}

// Unfold 展开地址段
func (ep Endpoint) Unfold() (Endpoints, error) {
	if !ep.IsCIDRAddr() {
		return nil, fmt.Errorf("only unfold cidr")
	}
	_, ipnet, _ := net.ParseCIDR(string(ep))
	segment := Endpoint(ipnet.IP.String())
	ipx, err := segment.formatUint32()
	if err != nil {
		return nil, err
	}
	eps := make(Endpoints, 0, 256)
	ones, bits := ipnet.Mask.Size()
	hostCount := 1<<uint(bits-ones) - 2
	for ipx, i := ipx+1, 0; i < hostCount; ipx, i = ipx+1, i+1 {
		var ep Endpoint
		err := (&ep).parseUint32(ipx)
		if err != nil {
			return nil, err
		}
		eps = append(eps, ep)
	}
	return eps, nil
}

// IsCIDRAddr 判断是不是地址段
func (ep Endpoint) IsCIDRAddr() bool {
	_, _, err := net.ParseCIDR(string(ep))
	if err != nil {
		return false
	}
	return true
}

// formatUint32 格式化到Uint32格式
func (ep Endpoint) formatUint32() (uint32, error) {
	if err := ep.Validate(); err != nil {
		return 0, err
	}
	if ep.IsCIDRAddr() {
		return 0, fmt.Errorf("can't format CIDR to uint32")
	}
	ip := net.ParseIP(string(ep))
	var ret uint32
	for i, v := range []byte(ip)[12:] {
		var temp uint32
		temp = temp | uint32(v)
		temp = temp << uint((3-i)*8)
		ret |= temp
	}
	return ret, nil
}

// parseUint32 解析uint32到Endpoint
func (ep *Endpoint) parseUint32(ipx uint32) error {
	data := make([]byte, 4)
	for j := 0; j < len(data); j++ {
		data[j] = byte(ipx >> uint((3-j)*8))
	}
	ip := net.IPv4(data[0], data[1], data[2], data[3])
	*ep = Endpoint(ip.String())
	return nil
}

// Endpoints 服务器地址列表
type Endpoints []Endpoint

// Validate 验证并且去掉不合法的数据
func (eps Endpoints) Validate() Endpoints {
	var validated Endpoints
	for _, ep := range eps {
		if err := ep.Validate(); err == nil {
			validated = append(validated, ep)
		}
	}
	return validated
}

// Unfold 展开整个列表 ****不会检查数据
func (eps Endpoints) Unfold() Endpoints {
	var unfolded Endpoints
	for _, ep := range eps {
		if !ep.IsCIDRAddr() {
			unfolded = append(unfolded, ep)
		} else {
			t, err := ep.Unfold()
			if err != nil {
				continue
			}
			unfolded = append(unfolded, t...)
		}
	}
	return unfolded
}

// Exclude 排除 ***不会自动展开****不会检查数据
func (eps Endpoints) Exclude(not Endpoints) Endpoints {
	var excluded Endpoints
	for _, ep := range eps {
		if !not.Contains(ep) {
			excluded = append(excluded, ep)
		}
	}
	return excluded
}

// Contains 是否包含  ***不会自动展开****不会检查数据
func (eps Endpoints) Contains(ep Endpoint) bool {
	for _, v := range eps {
		if ep.Equal(v) {
			return true
		}
	}
	return false
}

// ParseEndpoints 从字符串列表解析
func ParseEndpoints(rawData []string, allowCIDR bool) (Endpoints, error) {
	eps := make(Endpoints, 0, len(rawData))
	for _, rawEp := range rawData {
		ep := Endpoint(rawEp)
		if err := ep.Validate(); err != nil {
			return nil, err
		}
		if !allowCIDR && ep.IsCIDRAddr() {
			return nil, fmt.Errorf("cidr not allowed")
		}
		eps = append(eps, ep)
	}
	return eps, nil
}
