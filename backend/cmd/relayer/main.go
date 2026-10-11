// The relayer carries queued changes from the database to the chain. It runs
// beside the API and needs the same .env.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database config: %v", err)
	}
	defer pool.Close()

	// Dial does not need the chain to be up; an unreachable chain only makes items wait.
	client, err := ethclient.DialContext(ctx, cfg.ChainRPCURL)
	if err != nil {
		log.Fatalf("chain config: %v", err)
	}
	defer client.Close()

	signer, err := bind.NewKeyedTransactorWithChainID(cfg.PlatformKey, cfg.ChainID)
	if err != nil {
		log.Fatalf("platform key: %v", err)
	}
	contracts, err := chain.Bind(cfg.Contracts, client)
	if err != nil {
		log.Fatalf("contracts: %v", err)
	}

	log.Printf("relayer sending as %s on chain %s", signer.From.Hex(), cfg.ChainID)
	(&chain.Relayer{Pool: pool, Backend: client, Signer: signer, Contracts: contracts}).Run(ctx)
}
