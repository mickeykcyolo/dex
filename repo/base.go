package repo

import "database/sql"

// M is shorthand for map[string]interface{}
type M map[string]interface{}

// base is the base implementation of all repositories.
type base struct{ *sql.DB }

// baseTx is the base implementation of transactions in all repositories.
type baseTx struct{ *sql.Tx }

// Exec acts like sql.DB.Exec only it omits the result value.
func (b base) Exec(query string, args ...interface{}) (err error) {
	_, err = b.DB.Exec(query, args...)
	return
}

// Exec acts like sql.Tx.Exec only it omits the result value.
func (tx baseTx) Exec(query string, args ...interface{}) (err error) {
	_, err = tx.Tx.Exec(query, args...)
	return
}

// DoTransaction starts a new database transaction and calls doTx.
// If doTx returns without error, the transaction is committed,
// otherwise the transaction is rolled back.
func (b base) DoTransaction(doTx func(tx baseTx) error) (err error) {
	if tx, err := b.Begin(); err == nil {
		if err = doTx(baseTx{tx}); err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit()
	} else {
		return err
	}
}
