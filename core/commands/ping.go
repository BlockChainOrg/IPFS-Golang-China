
//此源码被清华学神尹成大魔王专业翻译分析并修改
//尹成QQ77025077
//尹成微信18510341407
//尹成所在QQ群721929980
//尹成邮箱 yinc13@mails.tsinghua.edu.cn
//尹成毕业于清华大学,微软区块链领域全球最有价值专家
//https://mvp.microsoft.com/zh-cn/PublicProfile/4033620
package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ipfs/go-ipfs/core/commands/cmdenv"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	pstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
	cmds "gx/ipfs/QmWGm4AbZEbnmdgVTza52MSNpEmBdFVqzmAysRbjrRyGbH/go-ipfs-cmds"
	"gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	iaddr "gx/ipfs/QmYDzHj9xwKN8gCXVJYxYBKxCwCwJURNkwgkvuPP69p3bX/go-ipfs-addr"
	ping "gx/ipfs/QmYxivS34F2M2n44WQQnRHGAKS8aoRUxwGpi9wk4Cdn4Jf/go-libp2p/p2p/protocol/ping"
	cmdkit "gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

const kPingTimeout = 10 * time.Second

type PingResult struct {
	Success bool
	Time    time.Duration
	Text    string
}

const (
	pingCountOptionName = "count"
)

//当用户尝试对自己执行ping操作时，将返回errpingself。
var ErrPingSelf = errors.New("error: can't ping self")

var PingCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Send echo request packets to IPFS hosts.",
		ShortDescription: `
'ipfs ping' is a tool to test sending data to other nodes. It finds nodes
via the routing system, sends pings, waits for pongs, and prints out round-
trip latency information.
		`,
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("peer ID", true, true, "ID of peer to be pinged.").EnableStdin(),
	},
	Options: []cmdkit.Option{
		cmdkit.IntOption(pingCountOptionName, "n", "Number of ping messages to send.").WithDefault(10),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

//必须在线！
		if !n.OnlineMode() {
			return ErrNotOnline
		}

		addr, pid, err := ParsePeerParam(req.Arguments[0])
		if err != nil {
			return fmt.Errorf("failed to parse peer address '%s': %s", req.Arguments[0], err)
		}

		if pid == n.Identity {
			return ErrPingSelf
		}

		if addr != nil {
n.Peerstore.AddAddr(pid, addr, pstore.TempAddrTTL) //暂时的
		}

		numPings, _ := req.Options[pingCountOptionName].(int)
		if numPings <= 0 {
			return fmt.Errorf("error: ping count must be greater than 0, was %d", numPings)
		}

		if len(n.Peerstore.Addrs(pid)) == 0 {
//确保我们可以找到有问题的节点
			if err := res.Emit(&PingResult{
				Text:    fmt.Sprintf("Looking up peer %s", pid.Pretty()),
				Success: true,
			}); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(req.Context, kPingTimeout)
			p, err := n.Routing.FindPeer(ctx, pid)
			cancel()
			if err != nil {
				return res.Emit(&PingResult{Text: fmt.Sprintf("Peer lookup error: %s", err)})
			}
			n.Peerstore.AddAddrs(p.ID, p.Addrs, pstore.TempAddrTTL)
		}

		if err := res.Emit(&PingResult{
			Text:    fmt.Sprintf("PING %s.", pid.Pretty()),
			Success: true,
		}); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(req.Context, kPingTimeout*time.Duration(numPings))
		defer cancel()
		pings, err := ping.Ping(ctx, n.PeerHost, pid)
		if err != nil {
			return res.Emit(&PingResult{
				Success: false,
				Text:    fmt.Sprintf("Ping error: %s", err),
			})
		}

		var total time.Duration
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for i := 0; i < numPings; i++ {
			t, ok := <-pings
			if !ok {
				break
			}

			if err := res.Emit(&PingResult{
				Success: true,
				Time:    t,
			}); err != nil {
				return err
			}

			total += t

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		averagems := total.Seconds() * 1000 / float64(numPings)
		return res.Emit(&PingResult{
			Success: true,
			Text:    fmt.Sprintf("Average latency: %.2fms", averagems),
		})
	},
	Type: PingResult{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, out *PingResult) error {
			if len(out.Text) > 0 {
				fmt.Fprintln(w, out.Text)
			} else if out.Success {
				fmt.Fprintf(w, "Pong received: time=%.2f ms\n", out.Time.Seconds()*1000)
			} else {
				fmt.Fprintf(w, "Pong failed\n")
			}
			return nil
		}),
	},
}

func ParsePeerParam(text string) (ma.Multiaddr, peer.ID, error) {
//多地址
	if strings.HasPrefix(text, "/") {
		a, err := iaddr.ParseString(text)
		if err != nil {
			return nil, "", err
		}
		return a.Transport(), a.ID(), nil
	}
//原始对等体ID
	p, err := peer.IDB58Decode(text)
	return nil, p, err
}
