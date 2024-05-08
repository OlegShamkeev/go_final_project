package main

type Task struct {
	Id      int    `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type Result struct {
	Id    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}
