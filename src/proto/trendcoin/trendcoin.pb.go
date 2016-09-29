// Code generated by protoc-gen-go.
// source: trendcoin.proto
// DO NOT EDIT!

/*
Package trendcoin is a generated protocol buffer package.

It is generated from these files:
	trendcoin.proto

It has these top-level messages:
	BalanceRequest
	BalanceReply
	TransactionData
	MakeTransactionsRequest
	MakeTransactionsReply
	TransactionLogRequest
	Transaction
	TransactionLogReply
*/
package trendcoin

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BalanceRequest struct {
	UserId uint64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *BalanceRequest) Reset()                    { *m = BalanceRequest{} }
func (m *BalanceRequest) String() string            { return proto.CompactTextString(m) }
func (*BalanceRequest) ProtoMessage()               {}
func (*BalanceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type BalanceReply struct {
	Balance int64  `protobuf:"varint,1,opt,name=balance" json:"balance,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
}

func (m *BalanceReply) Reset()                    { *m = BalanceReply{} }
func (m *BalanceReply) String() string            { return proto.CompactTextString(m) }
func (*BalanceReply) ProtoMessage()               {}
func (*BalanceReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type TransactionData struct {
	Source uint64 `protobuf:"varint,1,opt,name=source" json:"source,omitempty"`
	// if destination account do not exists, it will be created
	// be aware: there will be no checks for core user
	Destination uint64 `protobuf:"varint,2,opt,name=destination" json:"destination,omitempty"`
	Amount      uint64 `protobuf:"varint,3,opt,name=amount" json:"amount,omitempty"`
	Reason      string `protobuf:"bytes,4,opt,name=reason" json:"reason,omitempty"`
	// allows negative balance as a result
	AllowCredit bool `protobuf:"varint,5,opt,name=allow_credit,json=allowCredit" json:"allow_credit,omitempty"`
	// allows empty "source" or "destination" field
	AllowEmptySide bool `protobuf:"varint,6,opt,name=allow_empty_side,json=allowEmptySide" json:"allow_empty_side,omitempty"`
}

func (m *TransactionData) Reset()                    { *m = TransactionData{} }
func (m *TransactionData) String() string            { return proto.CompactTextString(m) }
func (*TransactionData) ProtoMessage()               {}
func (*TransactionData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type MakeTransactionsRequest struct {
	Transactions []*TransactionData `protobuf:"bytes,1,rep,name=transactions" json:"transactions,omitempty"`
}

func (m *MakeTransactionsRequest) Reset()                    { *m = MakeTransactionsRequest{} }
func (m *MakeTransactionsRequest) String() string            { return proto.CompactTextString(m) }
func (*MakeTransactionsRequest) ProtoMessage()               {}
func (*MakeTransactionsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MakeTransactionsRequest) GetTransactions() []*TransactionData {
	if m != nil {
		return m.Transactions
	}
	return nil
}

type MakeTransactionsReply struct {
	TransactionIds []uint64 `protobuf:"varint,1,rep,name=transaction_ids,json=transactionIds" json:"transaction_ids,omitempty"`
	Error          string   `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
}

func (m *MakeTransactionsReply) Reset()                    { *m = MakeTransactionsReply{} }
func (m *MakeTransactionsReply) String() string            { return proto.CompactTextString(m) }
func (*MakeTransactionsReply) ProtoMessage()               {}
func (*MakeTransactionsReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type TransactionLogRequest struct {
	UserId uint64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// default limit is 20
	Limit  uint64 `protobuf:"varint,2,opt,name=limit" json:"limit,omitempty"`
	Offset uint64 `protobuf:"varint,3,opt,name=offset" json:"offset,omitempty"`
	// created_at bounds, unixnano, [after, before)
	Before int64 `protobuf:"varint,4,opt,name=before" json:"before,omitempty"`
	After  int64 `protobuf:"varint,5,opt,name=after" json:"after,omitempty"`
}

func (m *TransactionLogRequest) Reset()                    { *m = TransactionLogRequest{} }
func (m *TransactionLogRequest) String() string            { return proto.CompactTextString(m) }
func (*TransactionLogRequest) ProtoMessage()               {}
func (*TransactionLogRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type Transaction struct {
	Id        uint64           `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	CreatedAt int64            `protobuf:"varint,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	Data      *TransactionData `protobuf:"bytes,3,opt,name=data" json:"data,omitempty"`
}

func (m *Transaction) Reset()                    { *m = Transaction{} }
func (m *Transaction) String() string            { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()               {}
func (*Transaction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *Transaction) GetData() *TransactionData {
	if m != nil {
		return m.Data
	}
	return nil
}

type TransactionLogReply struct {
	Transactions []*Transaction `protobuf:"bytes,1,rep,name=transactions" json:"transactions,omitempty"`
	Error        string         `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
}

func (m *TransactionLogReply) Reset()                    { *m = TransactionLogReply{} }
func (m *TransactionLogReply) String() string            { return proto.CompactTextString(m) }
func (*TransactionLogReply) ProtoMessage()               {}
func (*TransactionLogReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *TransactionLogReply) GetTransactions() []*Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func init() {
	proto.RegisterType((*BalanceRequest)(nil), "trendcoin.BalanceRequest")
	proto.RegisterType((*BalanceReply)(nil), "trendcoin.BalanceReply")
	proto.RegisterType((*TransactionData)(nil), "trendcoin.TransactionData")
	proto.RegisterType((*MakeTransactionsRequest)(nil), "trendcoin.MakeTransactionsRequest")
	proto.RegisterType((*MakeTransactionsReply)(nil), "trendcoin.MakeTransactionsReply")
	proto.RegisterType((*TransactionLogRequest)(nil), "trendcoin.TransactionLogRequest")
	proto.RegisterType((*Transaction)(nil), "trendcoin.Transaction")
	proto.RegisterType((*TransactionLogReply)(nil), "trendcoin.TransactionLogReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for TrendcoinService service

type TrendcoinServiceClient interface {
	Balance(ctx context.Context, in *BalanceRequest, opts ...grpc.CallOption) (*BalanceReply, error)
	// all requested transactions must end success or will be rollbacked
	MakeTransactions(ctx context.Context, in *MakeTransactionsRequest, opts ...grpc.CallOption) (*MakeTransactionsReply, error)
	TransactionLog(ctx context.Context, in *TransactionLogRequest, opts ...grpc.CallOption) (*TransactionLogReply, error)
}

type trendcoinServiceClient struct {
	cc *grpc.ClientConn
}

func NewTrendcoinServiceClient(cc *grpc.ClientConn) TrendcoinServiceClient {
	return &trendcoinServiceClient{cc}
}

func (c *trendcoinServiceClient) Balance(ctx context.Context, in *BalanceRequest, opts ...grpc.CallOption) (*BalanceReply, error) {
	out := new(BalanceReply)
	err := grpc.Invoke(ctx, "/trendcoin.TrendcoinService/Balance", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trendcoinServiceClient) MakeTransactions(ctx context.Context, in *MakeTransactionsRequest, opts ...grpc.CallOption) (*MakeTransactionsReply, error) {
	out := new(MakeTransactionsReply)
	err := grpc.Invoke(ctx, "/trendcoin.TrendcoinService/MakeTransactions", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trendcoinServiceClient) TransactionLog(ctx context.Context, in *TransactionLogRequest, opts ...grpc.CallOption) (*TransactionLogReply, error) {
	out := new(TransactionLogReply)
	err := grpc.Invoke(ctx, "/trendcoin.TrendcoinService/TransactionLog", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TrendcoinService service

type TrendcoinServiceServer interface {
	Balance(context.Context, *BalanceRequest) (*BalanceReply, error)
	// all requested transactions must end success or will be rollbacked
	MakeTransactions(context.Context, *MakeTransactionsRequest) (*MakeTransactionsReply, error)
	TransactionLog(context.Context, *TransactionLogRequest) (*TransactionLogReply, error)
}

func RegisterTrendcoinServiceServer(s *grpc.Server, srv TrendcoinServiceServer) {
	s.RegisterService(&_TrendcoinService_serviceDesc, srv)
}

func _TrendcoinService_Balance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrendcoinServiceServer).Balance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trendcoin.TrendcoinService/Balance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrendcoinServiceServer).Balance(ctx, req.(*BalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TrendcoinService_MakeTransactions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MakeTransactionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrendcoinServiceServer).MakeTransactions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trendcoin.TrendcoinService/MakeTransactions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrendcoinServiceServer).MakeTransactions(ctx, req.(*MakeTransactionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TrendcoinService_TransactionLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransactionLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrendcoinServiceServer).TransactionLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trendcoin.TrendcoinService/TransactionLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrendcoinServiceServer).TransactionLog(ctx, req.(*TransactionLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TrendcoinService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "trendcoin.TrendcoinService",
	HandlerType: (*TrendcoinServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Balance",
			Handler:    _TrendcoinService_Balance_Handler,
		},
		{
			MethodName: "MakeTransactions",
			Handler:    _TrendcoinService_MakeTransactions_Handler,
		},
		{
			MethodName: "TransactionLog",
			Handler:    _TrendcoinService_TransactionLog_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("trendcoin.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 498 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x54, 0x4d, 0x6f, 0x13, 0x31,
	0x10, 0x25, 0xdd, 0x6d, 0x42, 0x26, 0xd5, 0x36, 0x32, 0xb4, 0x59, 0x22, 0x81, 0x82, 0x2f, 0x84,
	0x4b, 0x0e, 0xe5, 0xc6, 0x01, 0xc4, 0xd7, 0xa1, 0x12, 0x5c, 0xdc, 0x08, 0x09, 0x2e, 0x91, 0x13,
	0x4f, 0xaa, 0x15, 0x9b, 0x75, 0xb0, 0x1d, 0x50, 0x7f, 0x00, 0x37, 0x7e, 0x18, 0x3f, 0x0b, 0x7f,
	0x6c, 0xb6, 0x9b, 0xb0, 0x0d, 0xb7, 0xbc, 0xb7, 0xcf, 0xcf, 0x33, 0x6f, 0xc6, 0x81, 0x53, 0xa3,
	0xb0, 0x10, 0x0b, 0x99, 0x15, 0x93, 0xb5, 0x92, 0x46, 0x92, 0x6e, 0x45, 0xd0, 0xe7, 0x90, 0xbc,
	0xe5, 0x39, 0x2f, 0x16, 0xc8, 0xf0, 0xfb, 0x06, 0xb5, 0x21, 0x03, 0xe8, 0x6c, 0x34, 0xaa, 0x59,
	0x26, 0xd2, 0xd6, 0xa8, 0x35, 0x8e, 0x59, 0xdb, 0xc1, 0x4b, 0x41, 0x5f, 0xc1, 0x49, 0x25, 0x5d,
	0xe7, 0x37, 0x24, 0x85, 0xce, 0x3c, 0x60, 0x2f, 0x8c, 0xd8, 0x16, 0x92, 0x87, 0x70, 0x8c, 0x4a,
	0x49, 0x95, 0x1e, 0x59, 0xbe, 0xcb, 0x02, 0xa0, 0x7f, 0x5a, 0x70, 0x3a, 0x55, 0xbc, 0xd0, 0x7c,
	0x61, 0x32, 0x59, 0xbc, 0xe7, 0x86, 0x93, 0x73, 0x68, 0x6b, 0xb9, 0x51, 0xa5, 0x85, 0xbd, 0x2b,
	0x20, 0x32, 0x82, 0x9e, 0xb0, 0xc5, 0x64, 0x05, 0x77, 0x52, 0xef, 0x13, 0xb3, 0x3a, 0xe5, 0x4e,
	0xf2, 0x95, 0xdc, 0x14, 0x26, 0x8d, 0xc2, 0xc9, 0x80, 0x1c, 0xaf, 0x90, 0x6b, 0x7b, 0x28, 0xf6,
	0x97, 0x97, 0x88, 0x3c, 0x85, 0x13, 0x9e, 0xe7, 0xf2, 0xe7, 0x6c, 0xa1, 0x50, 0x64, 0x26, 0x3d,
	0xb6, 0x5f, 0xef, 0xb3, 0x9e, 0xe7, 0xde, 0x79, 0x8a, 0x8c, 0xa1, 0x1f, 0x24, 0xb8, 0x5a, 0x9b,
	0x9b, 0x99, 0xce, 0x04, 0xa6, 0x6d, 0x2f, 0x4b, 0x3c, 0xff, 0xc1, 0xd1, 0x57, 0x96, 0xa5, 0x5f,
	0x60, 0xf0, 0x89, 0x7f, 0xc3, 0x5a, 0x37, 0x7a, 0x1b, 0x9f, 0x4d, 0xc9, 0xd4, 0x68, 0xdb, 0x57,
	0x34, 0xee, 0x5d, 0x0c, 0x27, 0xb7, 0x33, 0xd8, 0xcb, 0x80, 0xed, 0xe8, 0xe9, 0x67, 0x38, 0xfb,
	0xd7, 0xda, 0xc5, 0xfd, 0xcc, 0xcd, 0xb1, 0x22, 0xed, 0x78, 0x82, 0x77, 0xcc, 0x92, 0x1a, 0x7d,
	0x29, 0xf4, 0x1d, 0xe9, 0xff, 0x6e, 0xc1, 0x59, 0xcd, 0xf4, 0xa3, 0xbc, 0xfe, 0xdf, 0xc0, 0x9d,
	0x51, 0x9e, 0xad, 0x6c, 0x56, 0x21, 0xfe, 0x00, 0x5c, 0xc0, 0x72, 0xb9, 0xd4, 0x58, 0x05, 0x1f,
	0x90, 0xe3, 0xe7, 0xb8, 0x94, 0x0a, 0x7d, 0xf0, 0x11, 0x2b, 0x91, 0x73, 0xe1, 0x4b, 0x83, 0xca,
	0x27, 0x1e, 0xb1, 0x00, 0x68, 0x0e, 0xbd, 0x5a, 0x35, 0x24, 0x81, 0xa3, 0xea, 0x7a, 0xfb, 0x8b,
	0x3c, 0x06, 0xb0, 0x73, 0xe2, 0x06, 0xc5, 0x8c, 0x87, 0xfb, 0x23, 0xd6, 0x2d, 0x99, 0x37, 0x86,
	0x4c, 0x20, 0x16, 0x36, 0x3a, 0x5f, 0xc1, 0xe1, 0x70, 0xbd, 0x8e, 0x5e, 0xc3, 0x83, 0xfd, 0xde,
	0x5d, 0xa4, 0x2f, 0x1b, 0x67, 0x75, 0xde, 0x6c, 0xb7, 0x3b, 0xa7, 0xe6, 0x94, 0x2f, 0x7e, 0x1d,
	0x41, 0x7f, 0xba, 0x3d, 0x7d, 0x85, 0xea, 0x47, 0x66, 0x97, 0xf9, 0x35, 0x74, 0xca, 0x87, 0x43,
	0x1e, 0xd5, 0xbc, 0x77, 0xdf, 0xdd, 0x70, 0xd0, 0xf4, 0xc9, 0x56, 0x49, 0xef, 0x91, 0xaf, 0xd0,
	0xdf, 0xdf, 0x09, 0x42, 0x6b, 0xf2, 0x3b, 0x76, 0x71, 0x38, 0x3a, 0xa8, 0x09, 0xde, 0x53, 0x48,
	0x76, 0xa3, 0x21, 0xa3, 0xe6, 0xfe, 0x6f, 0x37, 0x66, 0xf8, 0xe4, 0x80, 0xc2, 0xbb, 0xce, 0xdb,
	0xfe, 0x8f, 0xe6, 0xc5, 0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x0c, 0xd1, 0xbe, 0x15, 0x7b, 0x04,
	0x00, 0x00,
}
