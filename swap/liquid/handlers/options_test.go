package handlers

import (
	"testing"

	"git.condensat.tech/bank/swap/liquid/common"
)

func TestLiquidSwapPropose(t *testing.T) {
	t.Parallel()

	proposal := common.ProposalInfo{
		ProposerAsset:  "assetP",
		ProposerAmount: 0.1234567811111,
		ReceiverAsset:  "assetR",
		ReceiverAmount: 3.141592653589793,
	}

	type args struct {
		address  common.ConfidentialAddress
		proposal common.ProposalInfo
		feeRate  float64
	}
	tests := []struct {
		name      string
		args      args
		wantEnv   int
		wantArgs  int
		wantStdIn bool
	}{
		{"propose", args{"address", proposal, 0.1337}, 2, 11, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := LiquidSwapPropose(tt.args.address, tt.args.proposal, tt.args.feeRate)

			if got.Program != LiquidSwapCli {
				t.Errorf("LiquidSwapPropose() wrong Program %v, want %v", got.Program, LiquidSwapCli)
			}
			if len(got.Env) != tt.wantEnv {
				t.Errorf("LiquidSwapPropose() Env = %v, want %v", len(got.Env), tt.wantEnv)
			}
			if len(got.Args) != tt.wantArgs {
				t.Errorf("LiquidSwapPropose() Args = %v, want %v", len(got.Args), tt.wantArgs)
			}
			if (got.Stdin != nil) != tt.wantStdIn {
				t.Errorf("LiquidSwapPropose() Stdin = %v, want %v", got.Stdin != nil, tt.wantStdIn)
			}

			t.Logf("Args: %v", got.Args)
		})
	}
}

func TestLiquidSwapInfo(t *testing.T) {
	t.Parallel()

	type args struct {
		payload common.Payload
	}
	tests := []struct {
		name      string
		args      args
		wantEnv   int
		wantArgs  int
		wantStdIn bool
	}{
		{"info", args{"payload"}, 2, 4, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := LiquidSwapInfo(tt.args.payload)

			if got.Program != LiquidSwapCli {
				t.Errorf("LiquidSwapInfo() wrong Program %v, want %v", got.Program, LiquidSwapCli)
			}
			if len(got.Env) != tt.wantEnv {
				t.Errorf("LiquidSwapInfo() Env = %v, want %v", len(got.Env), tt.wantEnv)
			}
			if len(got.Args) != tt.wantArgs {
				t.Errorf("LiquidSwapInfo() Args = %v, want %v", len(got.Args), tt.wantArgs)
			}
			if (got.Stdin != nil) != tt.wantStdIn {
				t.Errorf("LiquidSwapInfo() Stdin = %v, want %v", got.Stdin != nil, tt.wantStdIn)
			}

			t.Logf("Args: %v", got.Args)
		})
	}
}

func TestLiquidSwapFinalize(t *testing.T) {
	t.Parallel()

	type args struct {
		payload common.Payload
	}
	tests := []struct {
		name      string
		args      args
		wantEnv   int
		wantArgs  int
		wantStdIn bool
	}{
		{"finalize", args{"payload"}, 2, 5, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := LiquidSwapFinalize(tt.args.payload)

			if got.Program != LiquidSwapCli {
				t.Errorf("LiquidSwapFinalize() wrong Program %v, want %v", got.Program, LiquidSwapCli)
			}
			if len(got.Env) != tt.wantEnv {
				t.Errorf("LiquidSwapFinalize() Env = %v, want %v", len(got.Env), tt.wantEnv)
			}
			if len(got.Args) != tt.wantArgs {
				t.Errorf("LiquidSwapFinalize() Args = %v, want %v", len(got.Args), tt.wantArgs)
			}
			if (got.Stdin != nil) != tt.wantStdIn {
				t.Errorf("LiquidSwapFinalize() Stdin = %v, want %v", got.Stdin != nil, tt.wantStdIn)
			}

			t.Logf("Args: %v", got.Args)
		})
	}
}

func TestLiquidSwapAccept(t *testing.T) {
	t.Parallel()

	type args struct {
		address common.ConfidentialAddress
		payload common.Payload
		feeRate float64
	}
	tests := []struct {
		name      string
		args      args
		wantEnv   int
		wantArgs  int
		wantStdIn bool
	}{
		{"finalize", args{"address", "payload", 0.1337}, 2, 8, true},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := LiquidSwapAccept(tt.args.address, tt.args.payload, tt.args.feeRate)

			if got.Program != LiquidSwapCli {
				t.Errorf("LiquidSwapAccept() wrong Program %v, want %v", got.Program, LiquidSwapCli)
			}
			if len(got.Env) != tt.wantEnv {
				t.Errorf("LiquidSwapAccept() Env = %v, want %v", len(got.Env), tt.wantEnv)
			}
			if len(got.Args) != tt.wantArgs {
				t.Errorf("LiquidSwapAccept() Args = %v, want %v", len(got.Args), tt.wantArgs)
			}
			if (got.Stdin != nil) != tt.wantStdIn {
				t.Errorf("LiquidSwapAccept() Stdin = %v, want %v", got.Stdin != nil, tt.wantStdIn)
			}

			t.Logf("Args: %v", got.Args)
		})
	}
}
