/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import "time"

type Aml struct {
	Last_name  string `json:"last_name"`
	First_name string `json:"first_name"`
	DOB        string `json:"dob"`
	Country    string `json:"country"`
	ID_number  string `json:"id_number"`
	Data_owner string `json:"data_owner"`
	Risk_level string `json:"risk_level"`
}

type HistoryQueryResult struct {
	Record    *Aml      `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}
