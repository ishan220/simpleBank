package db

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	fmt.Printf("Type of store object returned from Newstore func %T", store)
	account1 := CreateRandomAccount(t)
	account2 := CreateRandomAccount(t)
	fmt.Println(">>Before Tx", account1.Balance, account2.Balance)
	//run n concurrrent transactions
	fmt.Println("store:", store)
	n := 10
	amount := int64(10)
	results := make(chan TransferTxResult, n)
	errs := make(chan error, n)
	//wg := &sync.WaitGroup{}
	//mut := &sync.RWMutex{}
	//wg.Add(n)

	for i := 0; i < n; i++ {
		txName := "tx" + strconv.Itoa(i+1)
		go func() {
			//strconv.ParseInt("-42", 10, 64)
			ctx := context.WithValue(context.Background(), txKey, txName)
			// result, err := store.TransferTx(ctx, TransferTxParams{
			// 	FromAccountID: account1.ID,
			// 	ToAccountID:   account2.ID,
			// 	Amount:        amount,
			// }, mut)
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			fmt.Println("Receiving to results channel:")
			results <- result
			fmt.Println("Receiving to Errors channel:")
			if err != nil {
				fmt.Println(txName, "This tx have some error")
			}
			errs <- err
			fmt.Println("Transaction Done  Inside Go Routine:")
			//wg.Done()

		}() ///anonymous function
	}
	//fmt.Println("Go Routines Created for the transactions")
	//wg.Wait()
	existed := make(map[int]bool)
	for j := 0; j < n; j++ {
		fmt.Println("LOG1")
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		FromEntry := result.FromEntry
		require.NotEmpty(t, FromEntry)

		require.Equal(t, account1.ID, FromEntry.AccountID)
		require.Equal(t, amount, FromEntry.Amount)
		require.NotZero(t, FromEntry.ID)
		require.NotZero(t, FromEntry.CreatedAt)

		_, err1 := store.GetEntry(context.Background(), FromEntry.ID)
		fmt.Println("Getting Entry for AccountId", FromEntry.ID)
		require.NoError(t, err1)

		ToEntry := result.ToEntry
		require.NotEmpty(t, ToEntry)

		require.Equal(t, account2.ID, ToEntry.AccountID)
		require.Equal(t, amount, ToEntry.Amount)
		require.NotZero(t, ToEntry.ID)
		require.NotZero(t, ToEntry.CreatedAt)

		_, err2 := store.GetEntry(context.Background(), ToEntry.ID)
		require.NoError(t, err2)

		///to check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		fmt.Println("LOG2")
		fmt.Println(">>For A Tx", fromAccount.Balance, toAccount.Balance)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance

		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) //diff1 = amount, amount*2, amount*3
		k := (int)(diff1 / amount)         //k will denote the transaction
		// responsible for deduction of amount from an account, as diff will be in multiple of amount
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	fmt.Println("LOG3")

	updatedAccount1, err := TestQueries.GetAccount(context.Background(), account1.ID)
	require.NotEqual(t, updatedAccount1, err)
	updatedAccount2, err := TestQueries.GetAccount(context.Background(), account2.ID)
	require.NotEqual(t, updatedAccount2, err)

	fmt.Println(">>After All Tx", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-amount*(int64)(n), updatedAccount1.Balance)
	require.Equal(t, account2.Balance+amount*(int64)(n), updatedAccount2.Balance)

}
