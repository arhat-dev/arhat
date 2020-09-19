// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: rpc.proto

// +build !mqtt,!coap

package aranyagopb

import (
	context "context"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("rpc.proto", fileDescriptor_77a6da22d6a3feb1) }

var fileDescriptor_77a6da22d6a3feb1 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2c, 0x2a, 0x48, 0xd6,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4b, 0x2c, 0x4a, 0xcc, 0xab, 0x4c, 0x94, 0xe2, 0x06,
	0x73, 0x21, 0x82, 0x46, 0x26, 0x5c, 0x3c, 0xce, 0xf9, 0x79, 0x79, 0xa9, 0xc9, 0x25, 0x99, 0x65,
	0x99, 0x25, 0x95, 0x42, 0x2a, 0x5c, 0x2c, 0xc1, 0x95, 0x79, 0xc9, 0x42, 0xdc, 0x7a, 0x10, 0xd5,
	0x7a, 0xbe, 0xc5, 0xe9, 0x52, 0x70, 0x8e, 0x73, 0x6e, 0x8a, 0x06, 0xa3, 0x01, 0xa3, 0x53, 0xf8,
	0x85, 0x87, 0x72, 0x0c, 0x37, 0x1e, 0xca, 0x31, 0x7c, 0x78, 0x28, 0xc7, 0xd8, 0xf0, 0x48, 0x8e,
	0x71, 0xc5, 0x23, 0x39, 0xc6, 0x13, 0x8f, 0xe4, 0x18, 0x2f, 0x3c, 0x92, 0x63, 0x7c, 0xf0, 0x48,
	0x8e, 0xf1, 0xc5, 0x23, 0x39, 0x86, 0x0f, 0x8f, 0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0,
	0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28, 0xc5, 0xc4, 0xa2, 0x8c, 0xc4, 0x12, 0xbd, 0x94,
	0xd4, 0x32, 0x7d, 0x88, 0x79, 0xba, 0x60, 0x37, 0x40, 0x39, 0xe9, 0xf9, 0x05, 0x49, 0x49, 0x6c,
	0x60, 0x11, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd7, 0x2c, 0x30, 0x8f, 0xb7, 0x00, 0x00,
	0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ConnectivityClient is the client API for Connectivity service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ConnectivityClient interface {
	Sync(ctx context.Context, opts ...grpc.CallOption) (Connectivity_SyncClient, error)
}

type connectivityClient struct {
	cc *grpc.ClientConn
}

func NewConnectivityClient(cc *grpc.ClientConn) ConnectivityClient {
	return &connectivityClient{cc}
}

func (c *connectivityClient) Sync(ctx context.Context, opts ...grpc.CallOption) (Connectivity_SyncClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Connectivity_serviceDesc.Streams[0], "/aranya.Connectivity/Sync", opts...)
	if err != nil {
		return nil, err
	}
	x := &connectivitySyncClient{stream}
	return x, nil
}

type Connectivity_SyncClient interface {
	Send(*Msg) error
	Recv() (*Cmd, error)
	grpc.ClientStream
}

type connectivitySyncClient struct {
	grpc.ClientStream
}

func (x *connectivitySyncClient) Send(m *Msg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *connectivitySyncClient) Recv() (*Cmd, error) {
	m := new(Cmd)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ConnectivityServer is the server API for Connectivity service.
type ConnectivityServer interface {
	Sync(Connectivity_SyncServer) error
}

// UnimplementedConnectivityServer can be embedded to have forward compatible implementations.
type UnimplementedConnectivityServer struct {
}

func (*UnimplementedConnectivityServer) Sync(srv Connectivity_SyncServer) error {
	return status.Errorf(codes.Unimplemented, "method Sync not implemented")
}

func RegisterConnectivityServer(s *grpc.Server, srv ConnectivityServer) {
	s.RegisterService(&_Connectivity_serviceDesc, srv)
}

func _Connectivity_Sync_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ConnectivityServer).Sync(&connectivitySyncServer{stream})
}

type Connectivity_SyncServer interface {
	Send(*Cmd) error
	Recv() (*Msg, error)
	grpc.ServerStream
}

type connectivitySyncServer struct {
	grpc.ServerStream
}

func (x *connectivitySyncServer) Send(m *Cmd) error {
	return x.ServerStream.SendMsg(m)
}

func (x *connectivitySyncServer) Recv() (*Msg, error) {
	m := new(Msg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Connectivity_serviceDesc = grpc.ServiceDesc{
	ServiceName: "aranya.Connectivity",
	HandlerType: (*ConnectivityServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Sync",
			Handler:       _Connectivity_Sync_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rpc.proto",
}