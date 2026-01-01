package domain

import "time"

type Item struct {
	DepartureTime string `json:"departure_time"`
	Count         any    `json:"count"`
	Price         string `json:"price"`
}

type Response struct {
	Items  []Item    `json:"items"`
	Moment time.Time `json:"moment"`
}
