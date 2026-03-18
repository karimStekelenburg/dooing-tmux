package ui

import "github.com/karimStekelenburg/dooing-tmux/internal/model"

func makeTodos(n int) []*model.Todo {
	todos := make([]*model.Todo, n)
	for i := range todos {
		todos[i] = model.NewTodo("test todo")
	}
	return todos
}
