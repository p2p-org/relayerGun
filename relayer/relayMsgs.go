package relayer

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RelayMsgs contains the msgs that need to be sent to both a src and dst chain
// after a given relay round. MaxTxSize and MaxMsgLength are ignored if they are
// set to zero.
type RelayMsgs struct {
	Src          []sdk.Msg
	Dst          []sdk.Msg
	MaxTxSize    uint64 // maximum permitted size of the msgs in a bundled relay transaction
	MaxMsgLength uint64 // maximum amount of messages in a bundled relay transaction

	last    bool
	success bool
}

// Ready returns true if there are messages to relay
func (r *RelayMsgs) Ready() bool {
	if r == nil {
		return false
	}

	if len(r.Src) == 0 && len(r.Dst) == 0 {
		return false
	}
	return true
}

// Success returns the success var
func (r *RelayMsgs) Success() bool {
	return r.success
}

func (r *RelayMsgs) IsMaxTx(msgLen, txSize uint64) bool {
	return (r.MaxMsgLength != 0 && msgLen > r.MaxMsgLength) ||
		(r.MaxTxSize != 0 && txSize > r.MaxTxSize)
}

// Send sends the messages with appropriate output
// TODO: Parallelize? Maybe?
func (r *RelayMsgs) Send(src, dst *Chain) {
	var msgLen, txSize uint64
	var msgs []sdk.Msg

	r.success = true

	time.Sleep(src.Delay)

	// submit batches of relay transactions
	for _, msg := range r.Src {
		msgLen++
		txSize += uint64(len(msg.GetSignBytes()))

		if r.IsMaxTx(msgLen, txSize) {
			// Submit the transactions to src chain and update its status
			r.success = r.success && send(src, msgs)

			// clear the current batch and reset variables
			msgLen, txSize = 1, uint64(len(msg.GetSignBytes()))
			msgs = []sdk.Msg{}
		}
		msgs = append(msgs, msg)
	}

	// submit leftover msgs
	if len(msgs) > 0 && !send(src, msgs) {
		r.success = false
	}

	// reset variables
	msgLen, txSize = 0, 0
	msgs = []sdk.Msg{}

	for _, msg := range r.Dst {
		msgLen++
		txSize += uint64(len(msg.GetSignBytes()))

		if r.IsMaxTx(msgLen, txSize) {
			// Submit the transaction to dst chain and update its status
			r.success = r.success && send(dst, msgs)

			// clear the current batch and reset variables
			msgLen, txSize = 1, uint64(len(msg.GetSignBytes()))
			msgs = []sdk.Msg{}
		}
		msgs = append(msgs, msg)
	}

	// submit leftover msgs
	if len(msgs) > 0 && !send(dst, msgs) {
		r.success = false
	}
}

// Submits the messages to the provided chain and logs the result of the transaction.
// Returns true upon success and false otherwise.
func send(chain *Chain, msgs []sdk.Msg) bool {
	res, err := chain.SendMsgs(msgs)
	if err != nil || res.Code != 0 {
		chain.LogFailedTx(res, err, msgs)
		return false
	} else {
		// NOTE: Add more data to this such as identifiers
		chain.LogSuccessTx(res, msgs)
	}
	return true
}

func (r *RelayMsgs) SendSync(src, dst *Chain) {
	var failed = false
	time.Sleep(src.Delay)
	// TODO: maybe figure out a better way to indicate error here?

	// TODO: Parallelize? Maybe?
	if len(r.Src) > 0 {
		// Submit the transactions to src chain
		res, err := src.SendMsgsSync(r.Src)
		if err != nil || res.Code != 0 {
			src.LogFailedTx(res, err, r.Src)
			failed = true
		} else {
			// NOTE: Add more data to this such as identifiers
			src.LogSuccessTx(res, r.Src)
		}
	}

	if len(r.Dst) > 0 {
		// Submit the transactions to dst chain
		res, err := dst.SendMsgsSync(r.Dst)
		if err != nil || res.Code != 0 {
			dst.LogFailedTx(res, err, r.Dst)
			failed = true
		} else {
			// NOTE: Add more data to this such as identifiers
			dst.LogSuccessTx(res, r.Dst)

		}
	}

	if failed {
		r.success = false
		return
	}
	r.success = true
}

func getMsgAction(msgs []sdk.Msg) string {
	var out string
	for i, msg := range msgs {
		out += fmt.Sprintf("%d:%s,", i, msg.Type())
	}
	return strings.TrimSuffix(out, ",")
}
