package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

var TX_TU *sqlx.Tx
var TX_VP *sqlx.Tx

func StartTransactionTU() {
	var err error
	TX_TU, err = DB_RT.Beginx()
	if err != nil {
		fmt.Println("StartTransaction db error:", err)
	}
}

func CommitTransactionTU() {
	err := TX_TU.Commit()
	if err != nil {
		fmt.Println("CommitTransaction db error:", err)
	}
}

func StartTransactionVP() {
	var err error
	TX_VP, err = DB_RT.Beginx()
	if err != nil {
		fmt.Println("StartTransaction db error:", err)
	}
}

func CommitTransactionVP() {
	err := TX_VP.Commit()
	if err != nil {
		fmt.Println("CommitTransaction db error:", err)
	}
}