/*
Copyright 2014 SAP SE

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"database/sql"
	"fmt"
	"testing"
)

func testTransactionCommit(db *sql.DB, t *testing.T) {
	table := RandomIdentifier("testTxCommit_")
	if _, err := db.Exec(fmt.Sprintf("create table %s (i tinyint)", table)); err != nil {
		t.Fatal(err)
	}

	tx1, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	tx2, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback()

	//insert record in transaction 1
	if _, err := tx1.Exec(fmt.Sprintf("insert into %s values(42)", table)); err != nil {
		t.Fatal(err)
	}

	//count records in transaction 1
	i := 0
	if err := tx1.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(fmt.Errorf("tx1: invalid number of records %d - 1 expected", i))
	}

	//count records in transaction 2 - isolation level 'read committed'' (default) expected, so no record should be there
	if err := tx2.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatal(fmt.Errorf("tx2: invalid number of records %d - 0 expected", i))
	}

	//commit insert
	if err := tx1.Commit(); err != nil {
		t.Fatal(err)
	}

	//in isolation level 'read commited' (default) record should be visible now
	if err := tx2.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(fmt.Errorf("tx2: invalid number of records %d - 1 expected", i))
	}
}

func testTransactionRollback(db *sql.DB, t *testing.T) {
	table := RandomIdentifier("testTxRollback_")
	if _, err := db.Exec(fmt.Sprintf("create table %s (i tinyint)", table)); err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	//insert record
	if _, err := tx.Exec(fmt.Sprintf("insert into %s values(42)", table)); err != nil {
		t.Fatal(err)
	}

	//count records
	i := 0
	if err := tx.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 1 {
		t.Fatal(fmt.Errorf("tx: invalid number of records %d - 1 expected", i))
	}

	//rollback insert
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	//new transaction
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	//rollback - no record expected
	if err := tx.QueryRow(fmt.Sprintf("select count(*) from %s", table)).Scan(&i); err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatal(fmt.Errorf("tx: invalid number of records %d - 0 expected", i))
	}
}

func TestTransaction(t *testing.T) {
	tests := []struct {
		name string
		fct  func(db *sql.DB, t *testing.T)
	}{
		{"transactionCommit", testTransactionCommit},
		{"transactionRollback", testTransactionRollback},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.fct(TestDB, t)
		})
	}
}
