package todov1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Todo struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListTodosRequest struct{}

type ListTodosResponse struct {
	Items []*Todo `json:"items"`
}

type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CreateTodoResponse struct {
	Item *Todo `json:"item"`
}

type TodoServiceServer interface {
	ListTodos(context.Context, *ListTodosRequest) (*ListTodosResponse, error)
	CreateTodo(context.Context, *CreateTodoRequest) (*CreateTodoResponse, error)
}

type UnimplementedTodoServiceServer struct{}

func (UnimplementedTodoServiceServer) ListTodos(context.Context, *ListTodosRequest) (*ListTodosResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListTodos not implemented")
}

func (UnimplementedTodoServiceServer) CreateTodo(context.Context, *CreateTodoRequest) (*CreateTodoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateTodo not implemented")
}

type TodoServiceClient interface {
	ListTodos(ctx context.Context, in *ListTodosRequest, opts ...grpc.CallOption) (*ListTodosResponse, error)
	CreateTodo(ctx context.Context, in *CreateTodoRequest, opts ...grpc.CallOption) (*CreateTodoResponse, error)
}

type todoServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTodoServiceClient(cc grpc.ClientConnInterface) TodoServiceClient {
	return &todoServiceClient{cc: cc}
}

func (c *todoServiceClient) ListTodos(ctx context.Context, in *ListTodosRequest, opts ...grpc.CallOption) (*ListTodosResponse, error) {
	out := new(ListTodosResponse)
	err := c.cc.Invoke(ctx, "/todo.v1.TodoService/ListTodos", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *todoServiceClient) CreateTodo(ctx context.Context, in *CreateTodoRequest, opts ...grpc.CallOption) (*CreateTodoResponse, error) {
	out := new(CreateTodoResponse)
	err := c.cc.Invoke(ctx, "/todo.v1.TodoService/CreateTodo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func RegisterTodoServiceServer(registrar grpc.ServiceRegistrar, srv TodoServiceServer) {
	registrar.RegisterService(&TodoService_ServiceDesc, srv)
}

var TodoService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "todo.v1.TodoService",
	HandlerType: (*TodoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListTodos",
			Handler:    _TodoService_ListTodos_Handler,
		},
		{
			MethodName: "CreateTodo",
			Handler:    _TodoService_CreateTodo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/todo/v1/todo.proto",
}

func _TodoService_ListTodos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTodosRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TodoServiceServer).ListTodos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/todo.v1.TodoService/ListTodos",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TodoServiceServer).ListTodos(ctx, req.(*ListTodosRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TodoService_CreateTodo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateTodoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TodoServiceServer).CreateTodo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/todo.v1.TodoService/CreateTodo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TodoServiceServer).CreateTodo(ctx, req.(*CreateTodoRequest))
	}
	return interceptor(ctx, in, info, handler)
}
