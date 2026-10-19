// The relayer carries queued changes from the database to the chain. It runs
// beside the API and needs the same .env.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/config"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
)

func main() {
	format, level := config.Logging()
	if _, err := logx.Setup(os.Stderr, format, level); err != nil {
		logx.Fatal("log settings", err)
	}
	cfg, err := config.Load()
	if err != nil {
		logx.Fatal("settings", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logx.Fatal("database config", err)
	}
	defer pool.Close()

	// Dial does not need the chain to be up; an unreachable chain only makes items wait.
	client, err := ethclient.DialContext(ctx, cfg.ChainRPCURL)
	if err != nil {
		logx.Fatal("chain config", err)
	}
	defer client.Close()

	signer, err := bind.NewKeyedTransactorWithChainID(cfg.RelayerKey, cfg.ChainID)
	if err != nil {
		logx.Fatal("relayer key", err)
	}
	contracts, err := chain.Bind(cfg.Contracts, client)
	if err != nil {
		logx.Fatal("contracts", err)
	}

	slog.Info("relayer started", "from", signer.From.Hex(), "chain_id", cfg.ChainID.String())
	(&chain.Relayer{Pool: pool, Backend: client, Signer: signer, Contracts: contracts}).Run(ctx)
}
