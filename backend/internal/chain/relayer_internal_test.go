package chain

import (
	"errors"
	"testing"
)

type rpcErr struct {
	msg  string
	data any
}

func (e rpcErr) Error() string  { return e.msg }
func (e rpcErr) ErrorData() any { return e.data }

// Only the contract refusing is final; node and wallet trouble is retried.
func TestRevertReason(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		refused bool
	}{
		{"unfunded sender", rpcErr{"gas required exceeds allowance (0)", nil}, false},
		{"insufficient funds", rpcErr{"insufficient funds for gas * price + value", nil}, false},
		{"revert without data", rpcErr{"execution reverted", nil}, true},
		{"revert with data", rpcErr{"execution reverted: custom error", "0x08c379a000000000"}, true},
		{"plain revert message", errors.New("execution reverted: Foo"), true},
		{"unreachable", errors.New("dial tcp: connection refused"), false},
	}
	for _, c := range cases {
		if _, got := revertReason(c.err); got != c.refused {
			t.Errorf("%s: refused = %v, want %v", c.name, got, c.refused)
		}
	}
}
