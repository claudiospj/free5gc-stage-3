package pfcp_udp

import (
	"net"
	"time"

	"free5gc/lib/pfcp"
	"free5gc/lib/pfcp/pfcpUdp"
	"free5gc/src/smf/logger"
	"free5gc/src/smf/smf_context"
	"free5gc/src/smf/smf_handler/smf_message"
	"free5gc/src/smf/smf_pfcp/pfcp_util"
)

const MaxPfcpUdpDataSize = 1024

var Server pfcpUdp.PfcpServer

var ServerStartTime time.Time

var SeqNumTable *pfcp_util.SeqNumTable

func init() {
	SeqNumTable = pfcp_util.NewSeqNumTable()
}

func Run() {
	CPNodeID := smf_context.SMF_Self().CPNodeID
	if len(CPNodeID.NodeIdValue) != 0 {
		Server.Addr = CPNodeID.ResolveNodeIdToIp().String()
	}
	err := Server.Listen()
	if err != nil {
		logger.PfcpLog.Errorf("Failed to listen: %v", err)
	}
	logger.PfcpLog.Infof("Listen on %s", Server.Conn.LocalAddr().String())

	go func(p *pfcpUdp.PfcpServer) {
		for {
			var pfcpMessage pfcp.Message
			remoteAddr, err := p.ReadFrom(&pfcpMessage)
			if err != nil {
				logger.PfcpLog.Errorf("Read PFCP error: %v", err)
				continue
			}

			seq_num_check_pass := SeqNumTable.RecvCheckAndPutItem(&pfcpMessage)
			if !seq_num_check_pass {
				logger.PfcpLog.Errorf("\nSequence Number checking error.\n")
				continue
			}

			pfcpUdpMessage := pfcpUdp.NewMessage(remoteAddr, &pfcpMessage)

			message := smf_message.NewPfcpMessage(&pfcpUdpMessage)
			smf_message.SendMessage(message)
		}
	}(&Server)

	ServerStartTime = time.Now()
}

func SendPfcp(msg pfcp.Message, addr *net.UDPAddr) {
	seq_num_check_pass := SeqNumTable.SendCheckAndPutItem(&msg)
	if !seq_num_check_pass {
		logger.PfcpLog.Errorf("\nSequence Number checking error.\n")
		return
	}

	err := Server.WriteTo(msg, addr)
	if err != nil {
		logger.PfcpLog.Errorf("Failed to send PFCP message: %v", err)
	}

}
