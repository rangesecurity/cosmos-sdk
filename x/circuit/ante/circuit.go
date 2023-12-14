package ante

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, blockTime time.Time, msgURL string, signers [][]byte) (bool, error)
}

// CircuitBreakerDecorator is an AnteDecorator that checks if the transaction type is allowed to enter the mempool or be executed
type CircuitBreakerDecorator struct {
	circuitKeeper CircuitBreaker
}

func NewCircuitBreakerDecorator(ck CircuitBreaker) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		circuitKeeper: ck,
	}
}

func (cbd CircuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// loop through all the messages and check if the message type is allowed
	for _, msg := range tx.GetMsgs() {
		msgURL := sdk.MsgTypeURL(msg)
		var signers [][]byte
		if lmsg, ok := msg.(sdk.LegacyMsg); ok {
			lmsgSigners := lmsg.GetSigners()
			for _, lmsgSigner := range lmsgSigners {
				signers = append(signers, lmsgSigner.Bytes())
			}
		} else {
			signatureMsg, ok := msg.(sdk.Signature)
			if !ok {
				return ctx, fmt.Errorf("message %s has no signer", msgURL)
			}
			signers = append(signers, signatureMsg.GetPubKey().Address().Bytes())
		}
		// todo: decide on course of action for when no signers are present
		if len(signers) == 0 {
			ctx.Logger().Error("recovered no signers", "msg.url", msgURL)
		}
		blockTime := ctx.BlockTime()
		isAllowed, err := cbd.circuitKeeper.IsAllowed(ctx, blockTime, msgURL, signers)
		if err != nil {
			return ctx, err
		}

		if !isAllowed {
			return ctx, errors.New("tx type not allowed")
		}
	}

	return next(ctx, tx, simulate)
}
