package data

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var todoType *graphql.Object
var userType *graphql.Object

var nodeDefinitions *relay.NodeDefinitions
var todosConnection *relay.GraphQLConnectionDefinitions

var Schema graphql.Schema

func init() {

	nodeDefinitions = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
		IDFetcher: func(id string, info graphql.ResolveInfo) interface{} {
			resolvedID := relay.FromGlobalID(id)
			if resolvedID.Type == "Todo" {
				return GetTodo(resolvedID.ID)
			}
			if resolvedID.Type == "User" {
				return GetUser(resolvedID.ID)
			}
			return nil
		},
		TypeResolve: func(value interface{}, info graphql.ResolveInfo) *graphql.Object {
			switch value.(type) {
			case *Todo:
				return todoType
			case *User:
				return userType
			default:
				return userType
			}
		},
	})

	todoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Todo",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("Todo", nil),
			"text": &graphql.Field{
				Type: graphql.String,
			},
			"complete": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
		Interfaces: []*graphql.Interface{nodeDefinitions.NodeInterface},
	})

	todosConnection = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Todo",
		NodeType: todoType,
	})

	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("User", nil),
			"todos": &graphql.Field{
				Type: todosConnection.ConnectionType,
				Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
					"status": &graphql.ArgumentConfig{
						Type:         graphql.String,
						DefaultValue: "any",
					},
				}),
				Resolve: func(p graphql.ResolveParams) interface{} {
					status, _ := p.Args["status"].(string)
					args := relay.NewConnectionArguments(p.Args)
					todos := TodosToSliceInterface(GetTodos(status))
					return relay.ConnectionFromArray(todos, args)
				},
			},
			"totalCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return len(GetTodos("any"))
				},
			},
			"completedCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return len(GetTodos("completed"))
				},
			},
		},
		Interfaces: []*graphql.Interface{nodeDefinitions.NodeInterface},
	})

	rootType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Root",
		Fields: graphql.Fields{
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
			"node": nodeDefinitions.NodeField,
		},
	})

	addTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "AddTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"text": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"todoEdge": &graphql.Field{
				Type: todosConnection.EdgeType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					return relay.EdgeType{
						Node:   todo,
						Cursor: relay.CursorForObjectInConnection(TodosToSliceInterface(GetTodos("any")), todo),
					}
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			text, _ := inputMap["text"].(string)
			todoId := AddTodo(text, false)
			return map[string]interface{}{
				"todoId": todoId,
			}
		},
	})

	changeTodoStatusMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "ChangeTodoStatus",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"complete": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
		},
		OutputFields: graphql.Fields{
			"todo": &graphql.Field{
				Type: todoType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					todo := GetTodo(todoId)
					return todo
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			complete, _ := inputMap["complete"].(bool)
			resolvedId := relay.FromGlobalID(id)
			ChangeTodoStatus(resolvedId.ID, complete)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}
		},
	})

	markAllTodosMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "MarkAllTodos",
		InputFields: graphql.InputObjectConfigFieldMap{
			"complete": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
		},
		OutputFields: graphql.Fields{
			"changedTodos": &graphql.Field{
				Type: graphql.NewList(todoType),
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoIds, _ := payload["todoIds"].([]string)
					todos := []*Todo{}
					for _, todoId := range todoIds {
						todo := GetTodo(todoId)
						if todo != nil {
							todos = append(todos, todo)
						}
					}
					return todos
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			complete, _ := inputMap["complete"].(bool)
			todoIds := MarkAllTodos(complete)
			return map[string]interface{}{
				"todoIds": todoIds,
			}
		},
	})

	removeCompletedTodosMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RemoveCompletedTodos",
		OutputFields: graphql.Fields{
			"deletedTodoIds": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoIds"]
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			todoIds := RemoveCompletedTodos()
			return map[string]interface{}{
				"todoIds": todoIds,
			}
		},
	})

	removeTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RemoveTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		OutputFields: graphql.Fields{
			"deletedTodoId": &graphql.Field{
				Type: graphql.ID,
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					return payload["todoId"]
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			resolvedId := relay.FromGlobalID(id)
			RemoveTodo(resolvedId.ID)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}
		},
	})
	renameTodoMutation := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "RenameTodo",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"text": &graphql.InputObjectFieldConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"todo": &graphql.Field{
				Type: todoType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					payload, _ := p.Source.(map[string]interface{})
					todoId, _ := payload["todoId"].(string)
					return GetTodo(todoId)
				},
			},
			"viewer": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) interface{} {
					return GetViewer()
				},
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo) map[string]interface{} {
			id, _ := inputMap["id"].(string)
			resolvedId := relay.FromGlobalID(id)
			text, _ := inputMap["text"].(string)
			RenameTodo(resolvedId.ID, text)
			return map[string]interface{}{
				"todoId": resolvedId.ID,
			}
		},
	})
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addTodo":              addTodoMutation,
			"changeTodoStatus":     changeTodoStatusMutation,
			"markAllTodos":         markAllTodosMutation,
			"removeCompletedTodos": removeCompletedTodosMutation,
			"removeTodo":           removeTodoMutation,
			"renameTodo":           renameTodoMutation,
		},
	})

	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootType,
		Mutation: mutationType,
	})
	if err != nil {
		panic(err)
	}
}
