package models

import (
	"fmt"
	"net"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func TestIPParse(t *testing.T) {
	ip1 := net.ParseIP("10.0.0.1")
	fmt.Println(ip1)
	ip2 := net.ParseIP("10.0.0")
	fmt.Println(ip2)
	ip3 := net.ParseIP("10.0.10.0")
	fmt.Println(ip3)
	ip4, ipn4, _ := net.ParseCIDR("10.0.0.255/24")
	fmt.Println(ip4, ipn4)
	fmt.Println(ipn4.Contains(ip3))
	fmt.Println(ipn4.Mask.Size())
	fmt.Println([]byte(ip1)[12:])
	var ipx uint32
	for i, v := range []byte(ip1)[12:] {
		var temp uint32
		temp = temp | uint32(v)
		temp = temp << uint((3-i)*8)
		fmt.Println(v, ",", temp)
		ipx |= temp
	}
	fmt.Println(ipx)
	count := 1<<(32-28) - 2
	for i := 0; i < count; i++ {
		newIpx := ipx + uint32(i)
		data := make([]byte, 4)
		for j := 0; j < len(data); j++ {
			data[j] = byte(newIpx >> uint((3-j)*8))
		}
		ip := net.IPv4(data[0], data[1], data[2], data[3])
		fmt.Println(ip)
	}
	fmt.Println(ip3.IsMulticast(), ip3.IsUnspecified(), ip3.IsGlobalUnicast())

	_, _, err := net.ParseCIDR("10.0.0.0")
	fmt.Println(err)
	ip5 := net.ParseIP("10.0.0.0/24")
	fmt.Println(ip5)
}

func TestEndpoint(t *testing.T) {
	assert.NoError(t, Endpoint("10.0.1.0").Validate())
	assert.Error(t, Endpoint("10.0.1.288").Validate())
	assert.True(t, Endpoint("10.0.0.1/26").IsCIDRAddr())
	assert.False(t, Endpoint("10.0.0.1").IsCIDRAddr())
	ipx, err := Endpoint("192.168.0.2").formatUint32()
	assert.NoError(t, err)
	assert.Equal(t, uint32(3232235522), ipx)
	var ep Endpoint
	assert.NoError(t, (&ep).parseUint32(3232235522))
	assert.True(t, ep.Equal(Endpoint("192.168.0.2")))
	eps, err := Endpoint("192.168.0.0/24").Unfold()
	assert.NoError(t, err)
	assert.Len(t, eps, 254)
	assert.True(t, eps[0].Equal(Endpoint("192.168.0.1")))
	assert.True(t, eps[len(eps)-1].Equal(Endpoint("192.168.0.254")))
	jsonData := `{"endpoints":["192.168.0.10","10.0.10.0/24"]}`
	type struct1 struct {
		Eps Endpoints `json:"endpoints"`
	}
	st := &struct1{}
	assert.NoError(t, json.Unmarshal([]byte(jsonData), st))
	assert.Equal(t, Endpoint("10.0.10.0/24"), st.Eps[1])
	data, err := json.Marshal(st)
	assert.NoError(t, err)
	assert.Equal(t, jsonData, string(data))
	unfolded := st.Eps.Unfold()
	assert.Len(t, unfolded, 255)
	excluded := unfolded.Exclude(Endpoints{Endpoint("10.0.10.1")})
	assert.False(t, excluded.Contains(Endpoint("10.0.10.1")))
}
