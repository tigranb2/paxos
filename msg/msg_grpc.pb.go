// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.17.3
// source: msg/msg.proto

package msg

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MessengerClient is the client API for Messenger service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessengerClient interface {
	MsgAcceptor(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error)
	MsgProposer(ctx context.Context, opts ...grpc.CallOption) (Messenger_MsgProposerClient, error)
	MsgLearner(ctx context.Context, in *SlotValue, opts ...grpc.CallOption) (*Empty, error)
}

type messengerClient struct {
	cc grpc.ClientConnInterface
}

func NewMessengerClient(cc grpc.ClientConnInterface) MessengerClient {
	return &messengerClient{cc}
}

func (c *messengerClient) MsgAcceptor(ctx context.Context, in *Msg, opts ...grpc.CallOption) (*Msg, error) {
	out := new(Msg)
	err := c.cc.Invoke(ctx, "/msg.Messenger/MsgAcceptor", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messengerClient) MsgProposer(ctx context.Context, opts ...grpc.CallOption) (Messenger_MsgProposerClient, error) {
	stream, err := c.cc.NewStream(ctx, &Messenger_ServiceDesc.Streams[0], "/msg.Messenger/MsgProposer", opts...)
	if err != nil {
		return nil, err
	}
	x := &messengerMsgProposerClient{stream}
	return x, nil
}

type Messenger_MsgProposerClient interface {
	Send(*QueueRequest) error
	Recv() (*SlotValue, error)
	grpc.ClientStream
}

type messengerMsgProposerClient struct {
	grpc.ClientStream
}

func (x *messengerMsgProposerClient) Send(m *QueueRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *messengerMsgProposerClient) Recv() (*SlotValue, error) {
	m := new(SlotValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *messengerClient) MsgLearner(ctx context.Context, in *SlotValue, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/msg.Messenger/MsgLearner", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessengerServer is the server API for Messenger service.
// All implementations must embed UnimplementedMessengerServer
// for forward compatibility
type MessengerServer interface {
	MsgAcceptor(context.Context, *Msg) (*Msg, error)
	MsgProposer(Messenger_MsgProposerServer) error
	MsgLearner(context.Context, *SlotValue) (*Empty, error)
	mustEmbedUnimplementedMessengerServer()
}

// UnimplementedMessengerServer must be embedded to have forward compatible implementations.
type UnimplementedMessengerServer struct {
}

func (UnimplementedMessengerServer) MsgAcceptor(context.Context, *Msg) (*Msg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MsgAcceptor not implemented")
}
func (UnimplementedMessengerServer) MsgProposer(Messenger_MsgProposerServer) error {
	return status.Errorf(codes.Unimplemented, "method MsgProposer not implemented")
}
func (UnimplementedMessengerServer) MsgLearner(context.Context, *SlotValue) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MsgLearner not implemented")
}
func (UnimplementedMessengerServer) mustEmbedUnimplementedMessengerServer() {}

// UnsafeMessengerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessengerServer will
// result in compilation errors.
type UnsafeMessengerServer interface {
	mustEmbedUnimplementedMessengerServer()
}

func RegisterMessengerServer(s grpc.ServiceRegistrar, srv MessengerServer) {
	s.RegisterService(&Messenger_ServiceDesc, srv)
}

func _Messenger_MsgAcceptor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Msg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessengerServer).MsgAcceptor(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/msg.Messenger/MsgAcceptor",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessengerServer).MsgAcceptor(ctx, req.(*Msg))
	}
	return interceptor(ctx, in, info, handler)
}

func _Messenger_MsgProposer_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MessengerServer).MsgProposer(&messengerMsgProposerServer{stream})
}

type Messenger_MsgProposerServer interface {
	Send(*SlotValue) error
	Recv() (*QueueRequest, error)
	grpc.ServerStream
}

type messengerMsgProposerServer struct {
	grpc.ServerStream
}

func (x *messengerMsgProposerServer) Send(m *SlotValue) error {
	return x.ServerStream.SendMsg(m)
}

func (x *messengerMsgProposerServer) Recv() (*QueueRequest, error) {
	m := new(QueueRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Messenger_MsgLearner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SlotValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessengerServer).MsgLearner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/msg.Messenger/MsgLearner",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessengerServer).MsgLearner(ctx, req.(*SlotValue))
	}
	return interceptor(ctx, in, info, handler)
}

// Messenger_ServiceDesc is the grpc.ServiceDesc for Messenger service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Messenger_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "msg.Messenger",
	HandlerType: (*MessengerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MsgAcceptor",
			Handler:    _Messenger_MsgAcceptor_Handler,
		},
		{
			MethodName: "MsgLearner",
			Handler:    _Messenger_MsgLearner_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "MsgProposer",
			Handler:       _Messenger_MsgProposer_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "msg/msg.proto",
}
