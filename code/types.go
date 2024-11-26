package main

import "sync"

type User struct {
	Name         string
	ID           int64
	MyTaskIDs    []int
	OwnerTaskIDs []int
}

type Task struct {
	Name    string
	byUser  User
	desUser User
	status  bool
}

type Handler struct {
	Users     map[int64]*User
	allTasks  map[int]*Task
	taskCount int
	commands  []string
	mu        *sync.Mutex
}
