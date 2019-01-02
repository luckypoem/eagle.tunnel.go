/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:42
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-02 15:27:28
 */

package eagletunnel

import (
	"strconv"
	"strings"
	"time"

	"../eaglelib/src"
)

// EtASK请求的类型
const (
	EtCheckUNKNOWN = iota
	EtCheckAuth
	EtCheckPING
	EtCheckVERSION
)

// ETCheck ET-Check协议的实现
type ETCheck struct {
}

// ParseEtCheckType 将字符串转换为EtCHECK请求的类型
func ParseEtCheckType(src string) int {
	switch src {
	case "AUTH", "auth":
		return EtCheckAuth
	case "PING", "ping":
		return EtCheckPING
	case "VERSION", "version":
		return EtCheckVERSION
	default:
		return EtCheckUNKNOWN
	}
}

// formatEtCheckType 得到EtCHECK请求类型对应的字符串
func formatEtCheckType(src int) string {
	switch src {
	case EtCheckAuth:
		return "AUTH"
	case EtCheckPING:
		return "PING"
	case EtCheckVERSION:
		return "VERSION"
	default:
		return "UNKNOWN"
	}
}

// Handle 处理ET-Check请求
func (ec *ETCheck) Handle(req Request, tunnel *eaglelib.Tunnel) {
	reqs := strings.Split(req.RequestMsgStr, " ")
	if len(reqs) < 2 {
		// 没有具体的Check请求内容
		return
	}
	theType := ParseEtCheckType(reqs[1])
	switch theType {
	case EtCheckPING:
		handleEtCheckPingReq(tunnel)
	case EtCheckVERSION:
		handleEtCheckVersionReq(tunnel, reqs)
	default:
	}
}

// SendEtCheckAuthReq 发射 ET-CHECK-AUTH 请求
func SendEtCheckAuthReq() string {
	// 当connect2Relayer成功，则说明鉴权成功
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
	if err != nil {
		return err.Error()
	}

	if LocalUser.ID == "root" {
		return "no local user"
	}
	return "AUTH OK with local user: " + LocalUser.ID
}

// SendEtCheckVersionReq 发射 ET-CHECK-VERSION 请求
func SendEtCheckVersionReq() string {
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
	if err != nil {
		return err.Error()
	}

	// 告知VERSION请求
	req := FormatEtType(EtCHECK) + " " +
		formatEtCheckType(EtCheckVERSION) + " " +
		ProtocolVersion.Raw
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return err.Error()
	}

	// 接受回复
	buffer := make([]byte, 1024)
	count, err := tunnel.ReadRight(buffer)
	if err != nil {
		return err.Error()
	}
	reply := string(buffer[:count])
	return reply
}

// SendEtCheckPingReq 发射ET-CHECK-PING请求
func SendEtCheckPingReq(sig interface{}) {
	sigChan := sig.(chan string)

	start := time.Now() // 开始计时

	// 连接服务器
	tunnel := eaglelib.Tunnel{}
	defer tunnel.Close()
	err := connect2Relayer(&tunnel)
	if err != nil {
		sigChan <- err.Error()
		return
	}

	// 告知PING请求
	req := FormatEtType(EtCHECK) + " " + formatEtCheckType(EtCheckPING)
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		sigChan <- err.Error()
		return
	}

	// 接收响应数据
	buffer := make([]byte, 8)
	count, err := tunnel.ReadRight(buffer)
	end := time.Now() // 停止计时
	if err != nil {
		sigChan <- err.Error()
		return
	}
	reply := string(buffer[:count])
	if reply != "ok" {
		sigChan <- "invalid ping reply: " + reply
		return
	}
	duration := end.Sub(start)
	ns := duration.Nanoseconds()
	ms := ns / 1000 / 1000
	sigChan <- strconv.FormatInt(ms, 10)
	return
}

func handleEtCheckPingReq(tunnel *eaglelib.Tunnel) {
	reply := "ok"
	tunnel.WriteLeft([]byte(reply))
}

func handleEtCheckVersionReq(tunnel *eaglelib.Tunnel, reqs []string) {
	if len(reqs) < 3 {
		reply := "no protocol version value"
		tunnel.WriteLeft([]byte(reply))
		return
	}
	versionOfReq, err := eaglelib.CreateVersion(reqs[2])
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return
	}
	if versionOfReq.IsLessThan(&ProtocolCompatibleVersion) {
		reply := "the version of protocol may be incompatible"
		tunnel.WriteLeft([]byte(reply))
		return
	}
	reply := "Protocol Version OK"
	tunnel.WriteLeft([]byte(reply))
}
