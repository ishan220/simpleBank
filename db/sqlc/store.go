package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Store provides all functions to execute db queries and trasaction
type Store struct { //example of composition analogy to inheritance in other languages
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

//BeginTx starts a transaction.

//The provided context is used until the transaction is committed or rolled back.
//If the context is canceled, the sql package will roll back the transaction.
// Tx.Commit will return an error if the context provided to BeginTx is canceled.
//The provided TxOptions is optional and may be nil if defaults should be used.
//If a non-default isolation level is used that the driver doesn't support,
//an error will be returned.

// type TxOptions struct {
// // Isolation is the transaction isolation level.
// //  If zero, the driver or database's default level is used.
//
//		Isolation IsolationLevel
//		ReadOnly  bool
//	}

//type DB struct {
// contains filtered or unexported fields
//}
//DB is a database handle representing a pool of zero or more underlying connections.
// It's safe for concurrent use by multiple goroutines.
//The sql package creates and frees connections automatically; it also maintains a
//free pool of idle connections. If the database has a concept of per-connection state
//,such state can be reliably observed within a transaction (Tx) or
//connection (Conn). Once DB.Begin is called, the returned Tx is bound to a
//single connection. Once Commit or Rollback is called on the transaction,
//that transaction's connection is returned to DB's idle connection pool.
//The pool size can be controlled with SetMaxIdleConns.

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	q := New(tx)
	err1 := fn(q)
	if err1 != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err:%v,rb err:%v", err1, rbErr)
		}
	}
	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_acccount_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"From Account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams, mut *sync.RWMutex) (TransferTxResult, error) {
	var result TransferTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		txName := ctx.Value(txKey)
		fmt.Println(txName, "Created Transfer")
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}
		fmt.Println("Transfer Created for The Transaction")

		fmt.Println(txName, "Created From Entry")

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}
		fmt.Println("FromEntry AccountId:", result.FromEntry.AccountID)

		//fmt.Println("FromEntry Created for The Transaction")

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		fmt.Println(txName, "Created To Entry")

		//fmt.Println("ToEntry Created for The Transaction")
		if err != nil {
			return err
		}
		fmt.Println("ToEntry AccountId:", result.ToEntry.AccountID)

		//updated account Balance
		//mut.Lock() //creating a lock on read and update as one go routine should
		//execute both the operations first
		//before the other routine get the access
		//The above lock will work if we use simple get sql query i.e without update
		fmt.Println(txName, "Get Account1 ")

		// account1, err := q.GetAccountForUpdate(ctx, arg.FromAccountID)
		// if err != nil {
		// 	//mut.Unlock()
		// 	return err
		// }
		// fmt.Println(txName, "Update Account1 ")

		result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			//	mut.Unlock()
			return err
		}
		//mut.Unlock()

		//fmt.Println("FromAccount FromAccount for The Transaction")
		fmt.Println("From Account Updated for The Transaction for Account:", result.FromAccount.ID)

		//mut.Lock()
		//fmt.Println(txName, "Get Account2 ")

		// account2, err := q.GetAccountForUpdate(ctx, arg.ToAccountID)
		// if err != nil {
		// 	//	mut.Unlock()
		// 	return err
		// }
		fmt.Println(txName, "Update Account1 ")

		result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.ToAccountID,
			Amount: arg.Amount,
		})

		if err != nil {
			//	mut.Unlock()
			return err
		}
		//mut.Unlock()

		//	fmt.Println("To Account Updated for The Transaction")
		fmt.Println("To Account Updated for The Transaction for Account:", result.ToAccount.ID)
		fmt.Println("result.FromEntry:", result.FromEntry)
		fmt.Println("result.ToEntry:", result.ToEntry)
		fmt.Println("result.Transfer:", result.Transfer)
		if err != nil {
			fmt.Println("Inside TransferTx,Error is not nil for txn", txName)
		}
		return err
	})

	return result, err
}
