package db

import (
	"context"
	"fmt"
	"testing"

	"SimpleBank/db/util"

	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, account1 Account, account2 Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomMoney(),
	}
	tnsfr, err := TestQueries.CreateTransfer(context.Background(), arg)

	if err != nil {
		fmt.Println("")
	}

	require.NoError(t, err)
	require.Equal(t, tnsfr.FromAccountID, arg.FromAccountID)
	require.Equal(t, tnsfr.ToAccountID, arg.ToAccountID)
	require.Equal(t, tnsfr.Amount, arg.Amount)
	return tnsfr
}

func TestCreateTransfer(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println("Account1 created for Transfer")
	fmt.Println("Account2 created for Transfer")
	createRandomTransfer(t, account1, account2)

}
func TestGetTransfer(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	transfer1 := createRandomTransfer(t, account1, account2)
	transfer2, err := TestQueries.GetTransfer(context.Background(), transfer1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, transfer2)
	require.Equal(t, transfer2.ID, transfer1.ID)
	require.Equal(t, transfer2.FromAccountID, transfer1.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)

}

func TestListTransfer(t *testing.T) {
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	for i := 0; i < 5; i++ {
		createRandomTransfer(t, account1, account2)
		createRandomTransfer(t, account2, account1)

	}
	arg := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        0,
	}
	allTransfers, err := TestQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, allTransfers, 5)
	for _, transfer := range allTransfers {
		require.Equal(t, transfer.FromAccountID, arg.FromAccountID)
		require.True(t, transfer.FromAccountID == account1.ID || transfer.ToAccountID == account1.ID)
	}
}
