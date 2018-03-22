package sofarpc

import (
	"fmt"
	"gitlab.alipay-inc.com/afe/mosn/pkg/types"
	"gitlab.alipay-inc.com/afe/mosn/pkg/log"
)

//All of the protocolMaps

var defaultProtocols = &protocols{
	protocolMaps: make(map[byte]Protocol),
}

type protocols struct {
	protocolMaps map[byte]Protocol
}

func DefaultProtocols() types.Protocols {
	return defaultProtocols
}

func NewProtocols(protocolMaps map[byte]Protocol) types.Protocols {
	return &protocols{
		protocolMaps: protocolMaps,
	}
}

func (p *protocols) Encode(value interface{}, data types.IoBuffer) {}

// filter = type.serverStreamConnection
func (p *protocols) Decode(data types.IoBuffer, filter types.DecodeFilter) {
	readableBytes := uint64(data.Len())

	//at least 1 byte for protocol code recognize
	if readableBytes > 1 {
		protocolCode := data.Bytes()[0]
		maybeProtocolVersion := data.Bytes()[1]

		log.DefaultLogger.Println("[Decoder]protocol code = ", protocolCode, ", maybeProtocolVersion = ", maybeProtocolVersion)

		if proto, exists := p.protocolMaps[protocolCode]; exists {
			var out = make([]RpcCommand, 0, 1)

			decoder := proto.GetDecoder()
			read := decoder.Decode(filter, data, &out)     //先解析称command,保存在OUT

			if len(out) > 0 {
				proto.GetCommandHandler().HandleCommand(filter, out[0])  //做decode 同时序列化，在此调用！！

				filter.OnDecodeComplete(out[0].GetId(), data.Cut(read))
			}
		} else {
			fmt.Println("Unknown protocol code: [", protocolCode, "] while decode in ProtocolDecoder.")
		}
	}
}

//TODO move this to seperate type 'ProtocolDecoer' or 'CodecEngine'
func (p *protocols) doDecode(ctx interface{}, data types.IoBuffer, out interface{}) {
	readableBytes := uint64(data.Len())

	//at least 1 byte for protocol code recognize
	if readableBytes > 1 {
		bytes := data.Bytes()
		protocolCode := bytes[0]
		maybeProtocolVersion := bytes[1]

		log.DefaultLogger.Println("[Decoder]protocol code = ", protocolCode, ", maybeProtocolVersion = ", maybeProtocolVersion)

		if proto, exists := p.protocolMaps[protocolCode]; exists {
			decoder := proto.GetDecoder()
			decoder.Decode(ctx, data, out)

			proto.GetCommandHandler().HandleCommand(ctx, decoder)

		} else {
			fmt.Println("Unknown protocol code: [", protocolCode, "] while decode in ProtocolDecoder.")
		}
	}
}

//TODO move this to seperate type 'ProtocolDecoer' or 'CodecEngine'
func (p *protocols) doHandle(protocolCode byte, ctx interface{}, msg interface{}) {
	if proto, exists := p.protocolMaps[protocolCode]; exists {
		proto.GetCommandHandler().HandleCommand(ctx, msg)
	} else {
		fmt.Println("Unknown protocol code: [", protocolCode, "] while doHandle in rpc handler.")
	}
}

func (p *protocols) RegisterProtocol(protocolCode byte, protocol Protocol) {
	if _, exists := p.protocolMaps[protocolCode]; exists {
		fmt.Println("Protocol alreay Exist:", protocolCode)
	} else {
		p.protocolMaps[protocolCode] = protocol
	}
}

func (p *protocols) UnRegisterProtocol(protocolCode byte) {
	if _, exists := p.protocolMaps[protocolCode]; exists {
		delete(p.protocolMaps, protocolCode)
		fmt.Println("Delete Protocol:", protocolCode)
	}
}

func RegisterProtocol(protocolCode byte, protocol Protocol) {
	defaultProtocols.RegisterProtocol(protocolCode, protocol)
}

func UnRegisterProtocol(protocolCode byte) {
	defaultProtocols.UnRegisterProtocol(protocolCode)
}
