/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

type Aml struct {
	Last_name  string `json:"last_name"`
	First_name string `json:"first_name"`
	DOB        string `json:"dob"`
	Country    string `json:"country"`
	ID_number  string `json:"id_number"`
	Data_owner string `json:"data_owner"`
	Risk_level string `json:"risk_level"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Tracking_number string `json:"tracking_number"`
	Record          *Aml
}
