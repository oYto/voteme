package graphql

import (
	"VoteMe/db"    // 导入db包用于数据库操作
	"VoteMe/utils" // 导入utils包用于获取当前票据
	"fmt"
	"github.com/graphql-go/graphql" // 导入graphql包用于创建GraphQL服务
)

// 定义GraphQL中的用户类型
// 这个类型包含两个字段：name和votes，分别表示用户名和票数
var userType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "User", // 类型的名字
		Fields: graphql.Fields{ // 字段定义
			"name": &graphql.Field{
				Type: graphql.String, // 字段类型为字符串
			},
			"votes": &graphql.Field{
				Type: graphql.Int, // 票数字段类型为整数
			},
		},
	},
)

// 定义GraphQL中的票据类型
// ticketID和validity，分别表示票据ID和其有效性
var ticketType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Ticket", // 类型的名字
		Fields: graphql.Fields{ // 字段定义
			"ticketID": &graphql.Field{
				Type: graphql.String, // 票据ID字段类型为字符串
			},
			"validity": &graphql.Field{
				Type: graphql.Boolean, // 有效性字段类型为布尔值
			},
		},
	},
)

// 定义GraphQL查询类型
// 这里定义了两个查询：getUserVotes和getCurrentTicket
var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"getUserVotes": &graphql.Field{
				Type: graphql.Int, // 返回类型为整数，直接返回票数
				Args: graphql.FieldConfigArgument{ // 查询参数
					"name": &graphql.ArgumentConfig{
						Type: graphql.String, // 参数类型为字符串
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) { // 解析函数
					name, _ := params.Args["name"].(string)
					votes, err := db.GetUserVotes(name) // 调用数据库操作获取票数
					if err != nil {
						return nil, fmt.Errorf("error getting votes for user %s: %s", name, err)
					}
					return votes, nil
				},
			},
			"getCurrentTicket": &graphql.Field{ // 获取当前票据查询
				Type: ticketType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					currentTicket := utils.GetCurrentTicket() // 获取当前票据
					return map[string]interface{}{
						"ticketID": currentTicket,
						"validity": true,
					}, nil
				},
			},
		},
	},
)

// 定义GraphQL变更类型
// 这里定义了一个变更操作：vote
var mutationType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"vote": &graphql.Field{
				Type: graphql.Boolean, // 投票操作的返回类型为布尔值，表示是否成功
				Args: graphql.FieldConfigArgument{ // 变更参数
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewList(graphql.String), // 支持输入多个用户名
					},
					"ticket": &graphql.ArgumentConfig{
						Type: graphql.String, // 票据字段
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) { // 解析函数
					//ticket, _ := params.Args["ticket"].(string)
					names, _ := params.Args["name"].([]interface{})
					//
					//// 检查票据是否合法
					//if ticket != utils.GetCurrentTicket() {
					//	return false, fmt.Errorf("invalid or expired ticket")
					//}

					ticketID, _ := params.Args["ticket"].(string)
					_, err := db.CreateOrUpdateTicket(ticketID)
					if err != nil {
						return false, err
					}

					// 对每个用户名执行投票操作
					for _, nameInterface := range names {
						name, ok := nameInterface.(string)
						if !ok {
							return false, fmt.Errorf("invalid name type")
						}
						err := db.UpdateUserVotesWithLock(name) // 调用数据库操作增加票数
						if err != nil {
							return false, err
						}
					}

					return true, nil // 如果所有操作成功，返回true
				},
			},
		},
	},
)

// NewGraphQLSchema 创建新的GraphQL schema
// 这个函数将上面定义的查询类型和变更类型组合成一个完整的schema
func NewGraphQLSchema() (graphql.Schema, error) {
	Schema, err := graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)
	return Schema, err
}
