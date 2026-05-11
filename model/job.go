package model

type Job struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Company   string `json:"company"`
	City	  string `json:"city"`
}