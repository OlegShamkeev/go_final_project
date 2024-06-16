package main

type Task struct {
	Id      string `json:"id,omitempty" db:"id"`
	Date    string `json:"date,omitempty" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment,omitempty" db:"comment"`
	Repeat  string `json:"repeat,omitempty" db:"repeat"`
}

type Result struct {
	Id    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}
