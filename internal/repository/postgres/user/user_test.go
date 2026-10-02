package user

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/gr0shka/Tic-tac-toe/internal/repository/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, cleanup, err := testutil.SetupTestDB(ctx)
	if err != nil {
		log.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()

	testPool = pool
	os.Exit(m.Run())
}
