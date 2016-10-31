// Code generated by protoc-gen-go.
// source: payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment.proto

It has these top-level messages:
	CreateOrderRequest
	PaymentNotification
	CreateOrderReply
	BuyOrderRequest
	BuyOrderReply
	CancelOrderRequest
	CancelOrderReply
	ChatMessageNewOrder
	ChatMessagePaymentFinished
	ChatMessageOrderCancelled
*/
package payment

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

type Currency int32

const (
	Currency_RUB Currency = 0
	Currency_USD Currency = 1
	Currency_COP Currency = 2
)

var Currency_name = map[int32]string{
	0: "RUB",
	1: "USD",
	2: "COP",
}
var Currency_value = map[string]int32{
	"RUB": 0,
	"USD": 1,
	"COP": 2,
}

func (x Currency) String() string {
	return proto.EnumName(Currency_name, int32(x))
}
func (Currency) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Errors int32

const (
	Errors_OK Errors = 0
	// internal errors
	Errors_INVALID_DATA       Errors = 1
	Errors_DB_FAILED          Errors = 2
	Errors_ALREADY_PAYED      Errors = 3
	Errors_PAY_CANCELLED      Errors = 4
	Errors_ANOTHER_OPEN_ORDER Errors = 5
	Errors_UNKNOWN_ERROR      Errors = 126
	// external errors
	Errors_INIT_FAILED Errors = 127
	Errors_PAY_FAILED  Errors = 128
	Errors_CHAT_DOWN   Errors = 129
	Errors_COINS_DOWN  Errors = 130
	// commission source lacks funds
	Errors_CANT_PAY_FEE Errors = 131
	// refund fails
	// realy bad in case of Create order: commission was writed off, but pay wasn't created
	Errors_REFUND_ERROR Errors = 132
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	3:   "ALREADY_PAYED",
	4:   "PAY_CANCELLED",
	5:   "ANOTHER_OPEN_ORDER",
	126: "UNKNOWN_ERROR",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
	129: "CHAT_DOWN",
	130: "COINS_DOWN",
	131: "CANT_PAY_FEE",
	132: "REFUND_ERROR",
}
var Errors_value = map[string]int32{
	"OK":                 0,
	"INVALID_DATA":       1,
	"DB_FAILED":          2,
	"ALREADY_PAYED":      3,
	"PAY_CANCELLED":      4,
	"ANOTHER_OPEN_ORDER": 5,
	"UNKNOWN_ERROR":      126,
	"INIT_FAILED":        127,
	"PAY_FAILED":         128,
	"CHAT_DOWN":          129,
	"COINS_DOWN":         130,
	"CANT_PAY_FEE":       131,
	"REFUND_ERROR":       132,
}

func (x Errors) String() string {
	return proto.EnumName(Errors_name, int32(x))
}
func (Errors) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Direction int32

const (
	Direction_CLIENT_PAYS Direction = 0
	Direction_CLIENT_RECV Direction = 1
)

var Direction_name = map[int32]string{
	0: "CLIENT_PAYS",
	1: "CLIENT_RECV",
}
var Direction_value = map[string]int32{
	"CLIENT_PAYS": 0,
	"CLIENT_RECV": 1,
}

func (x Direction) String() string {
	return proto.EnumName(Direction_name, int32(x))
}
func (Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type CreateOrderRequest struct {
	Amount         uint64    `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency       Currency  `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId         uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	UserId         uint64    `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ConversationId uint64    `protobuf:"varint,5,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	Direction      Direction `protobuf:"varint,6,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ShopCardNumber string    `protobuf:"bytes,7,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,8,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,9,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
	// ewallet parameters
	PaymentMethod string `protobuf:"bytes,10,opt,name=payment_method,json=paymentMethod" json:"payment_method,omitempty"`
	Comment       string `protobuf:"bytes,11,opt,name=comment" json:"comment,omitempty"`
	ServiceId     string `protobuf:"bytes,12,opt,name=service_id,json=serviceId" json:"service_id,omitempty"`
	CardId        string `protobuf:"bytes,13,opt,name=card_id,json=cardId" json:"card_id,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type PaymentNotification struct {
	Id       uint64   `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	LeadId   uint64   `protobuf:"varint,4,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	UserId   uint64   `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,6,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,7,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
	// ewallet parameters
	PaymentMethod string `protobuf:"bytes,8,opt,name=payment_method,json=paymentMethod" json:"payment_method,omitempty"`
	Comment       string `protobuf:"bytes,9,opt,name=comment" json:"comment,omitempty"`
	ServiceId     string `protobuf:"bytes,10,opt,name=service_id,json=serviceId" json:"service_id,omitempty"`
	CardId        string `protobuf:"bytes,11,opt,name=card_id,json=cardId" json:"card_id,omitempty"`
}

func (m *PaymentNotification) Reset()                    { *m = PaymentNotification{} }
func (m *PaymentNotification) String() string            { return proto.CompactTextString(m) }
func (*PaymentNotification) ProtoMessage()               {}
func (*PaymentNotification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CreateOrderReply struct {
	Id           uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type BuyOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
	// parameters to check permissions
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type BuyOrderReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CancelOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	// parameters to check permissions
	LeadId    uint64    `protobuf:"varint,3,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	Direction Direction `protobuf:"varint,4,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	// userID just to log it
	UserId uint64 `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *CancelOrderRequest) Reset()                    { *m = CancelOrderRequest{} }
func (m *CancelOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderRequest) ProtoMessage()               {}
func (*CancelOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type CancelOrderReply struct {
	Cancelled    bool   `protobuf:"varint,1,opt,name=cancelled" json:"cancelled,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CancelOrderReply) Reset()                    { *m = CancelOrderReply{} }
func (m *CancelOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderReply) ProtoMessage()               {}
func (*CancelOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// chat messages
type ChatMessageNewOrder struct {
	PayId    uint64   `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type ChatMessagePaymentFinished struct {
	PayId     uint64    `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Success   bool      `protobuf:"varint,3,opt,name=success" json:"success,omitempty"`
	Failure   bool      `protobuf:"varint,4,opt,name=failure" json:"failure,omitempty"`
	Amount    uint64    `protobuf:"varint,5,opt,name=amount" json:"amount,omitempty"`
	Currency  Currency  `protobuf:"varint,6,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	Direction Direction `protobuf:"varint,7,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
}

func (m *ChatMessagePaymentFinished) Reset()                    { *m = ChatMessagePaymentFinished{} }
func (m *ChatMessagePaymentFinished) String() string            { return proto.CompactTextString(m) }
func (*ChatMessagePaymentFinished) ProtoMessage()               {}
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type ChatMessageOrderCancelled struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *ChatMessageOrderCancelled) Reset()                    { *m = ChatMessageOrderCancelled{} }
func (m *ChatMessageOrderCancelled) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageOrderCancelled) ProtoMessage()               {}
func (*ChatMessageOrderCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func init() {
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*PaymentNotification)(nil), "payment.PaymentNotification")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*CancelOrderRequest)(nil), "payment.CancelOrderRequest")
	proto.RegisterType((*CancelOrderReply)(nil), "payment.CancelOrderReply")
	proto.RegisterType((*ChatMessageNewOrder)(nil), "payment.ChatMessageNewOrder")
	proto.RegisterType((*ChatMessagePaymentFinished)(nil), "payment.ChatMessagePaymentFinished")
	proto.RegisterType((*ChatMessageOrderCancelled)(nil), "payment.ChatMessageOrderCancelled")
	proto.RegisterEnum("payment.Currency", Currency_name, Currency_value)
	proto.RegisterEnum("payment.Errors", Errors_name, Errors_value)
	proto.RegisterEnum("payment.Direction", Direction_name, Direction_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for PaymentService service

type PaymentServiceClient interface {
	CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderReply, error)
	BuyOrder(ctx context.Context, in *BuyOrderRequest, opts ...grpc.CallOption) (*BuyOrderReply, error)
	CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderReply, error)
}

type paymentServiceClient struct {
	cc *grpc.ClientConn
}

func NewPaymentServiceClient(cc *grpc.ClientConn) PaymentServiceClient {
	return &paymentServiceClient{cc}
}

func (c *paymentServiceClient) CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderReply, error) {
	out := new(CreateOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/CreateOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) BuyOrder(ctx context.Context, in *BuyOrderRequest, opts ...grpc.CallOption) (*BuyOrderReply, error) {
	out := new(BuyOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/BuyOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) CancelOrder(ctx context.Context, in *CancelOrderRequest, opts ...grpc.CallOption) (*CancelOrderReply, error) {
	out := new(CancelOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/CancelOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PaymentService service

type PaymentServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderReply, error)
	BuyOrder(context.Context, *BuyOrderRequest) (*BuyOrderReply, error)
	CancelOrder(context.Context, *CancelOrderRequest) (*CancelOrderReply, error)
}

func RegisterPaymentServiceServer(s *grpc.Server, srv PaymentServiceServer) {
	s.RegisterService(&_PaymentService_serviceDesc, srv)
}

func _PaymentService_CreateOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CreateOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/CreateOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CreateOrder(ctx, req.(*CreateOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_BuyOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BuyOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).BuyOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/BuyOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).BuyOrder(ctx, req.(*BuyOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_CancelOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).CancelOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/CancelOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).CancelOrder(ctx, req.(*CancelOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PaymentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler:    _PaymentService_CreateOrder_Handler,
		},
		{
			MethodName: "BuyOrder",
			Handler:    _PaymentService_BuyOrder_Handler,
		},
		{
			MethodName: "CancelOrder",
			Handler:    _PaymentService_CancelOrder_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 922 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x56, 0x4b, 0x6f, 0xdb, 0x46,
	0x10, 0x36, 0xf5, 0xe6, 0x58, 0x92, 0xe9, 0x0d, 0x9a, 0x32, 0x6e, 0x0b, 0xb4, 0x2a, 0x8c, 0x06,
	0x2e, 0x12, 0x14, 0xe9, 0xbd, 0x00, 0x4d, 0xd2, 0x8d, 0x60, 0x85, 0x32, 0x56, 0x52, 0x8a, 0x9c,
	0x08, 0x86, 0x5a, 0xd7, 0x04, 0x24, 0x52, 0x5d, 0x92, 0x29, 0x0c, 0x14, 0xe9, 0x23, 0xbd, 0xf7,
	0x27, 0xb6, 0xe8, 0x35, 0x7f, 0xa2, 0xb3, 0x4b, 0x52, 0xa4, 0xe2, 0x28, 0x56, 0x01, 0xdf, 0x38,
	0xdf, 0xcc, 0xee, 0x7c, 0xfb, 0xcd, 0x03, 0x84, 0xde, 0xca, 0xbb, 0x5e, 0xb2, 0x30, 0x79, 0xbc,
	0xe2, 0x51, 0x12, 0x91, 0x76, 0x6e, 0x0e, 0xfe, 0xae, 0x03, 0x31, 0x39, 0xf3, 0x12, 0x36, 0xe6,
	0x73, 0xc6, 0x29, 0xfb, 0x29, 0x65, 0x71, 0x42, 0xee, 0x43, 0xcb, 0x5b, 0x46, 0x69, 0x98, 0xe8,
	0xca, 0xe7, 0xca, 0xc3, 0x06, 0xcd, 0x2d, 0xf2, 0x08, 0x3a, 0x7e, 0xca, 0x39, 0x0b, 0xfd, 0x6b,
	0xbd, 0x86, 0x9e, 0xfe, 0x93, 0xc3, 0xc7, 0xc5, 0xcd, 0x66, 0xee, 0xa0, 0xeb, 0x10, 0xf2, 0x31,
	0xb4, 0x17, 0xcc, 0x9b, 0xbb, 0xc1, 0x5c, 0xaf, 0x67, 0xf7, 0x08, 0x73, 0x38, 0x17, 0x8e, 0x34,
	0x66, 0x5c, 0x38, 0x1a, 0x99, 0x43, 0x98, 0xe8, 0xf8, 0x0a, 0x0e, 0xfc, 0x28, 0x7c, 0xc5, 0x78,
	0xec, 0x25, 0x41, 0x14, 0x8a, 0x80, 0xa6, 0x0c, 0xe8, 0x57, 0x61, 0x0c, 0xfc, 0x06, 0xd4, 0x79,
	0xc0, 0x99, 0x2f, 0x4c, 0xbd, 0x25, 0xa9, 0x90, 0x35, 0x15, 0xab, 0xf0, 0xd0, 0x32, 0x88, 0x3c,
	0x04, 0x2d, 0xbe, 0x8a, 0x56, 0xae, 0xef, 0xf1, 0xb9, 0x1b, 0xa6, 0xcb, 0x97, 0x8c, 0xeb, 0x6d,
	0x3c, 0xa8, 0xd2, 0xbe, 0xc0, 0x4d, 0x84, 0x1d, 0x89, 0x92, 0x63, 0xc0, 0x6c, 0xcb, 0x65, 0x10,
	0xc7, 0x82, 0xc2, 0x25, 0x63, 0x7a, 0x47, 0x72, 0xe8, 0x95, 0xe8, 0x19, 0x63, 0xe4, 0x6b, 0x38,
	0xac, 0x84, 0xc5, 0x51, 0xca, 0x7d, 0xa6, 0xab, 0x32, 0x52, 0x2b, 0x1d, 0x13, 0x89, 0x8b, 0x3b,
	0x73, 0x76, 0xee, 0x92, 0x25, 0x57, 0xd1, 0x5c, 0x07, 0x99, 0xbb, 0x28, 0xcc, 0x33, 0x09, 0x12,
	0x1d, 0xda, 0xe2, 0x28, 0x02, 0xfa, 0xbe, 0xf4, 0x17, 0x26, 0xf9, 0x0c, 0x00, 0x25, 0x7a, 0x15,
	0xf8, 0x4c, 0x88, 0xd2, 0x95, 0x4e, 0x35, 0x47, 0x32, 0x45, 0xe5, 0xc3, 0xd0, 0xd7, 0x93, 0xbe,
	0x96, 0x30, 0x87, 0xf3, 0xc1, 0xdb, 0x1a, 0xdc, 0xbb, 0xc8, 0x72, 0x38, 0x51, 0x12, 0x5c, 0x06,
	0xbe, 0x94, 0x90, 0xf4, 0xa1, 0x86, 0xb1, 0x59, 0x79, 0xf1, 0xab, 0x52, 0xf2, 0xda, 0xd6, 0x92,
	0xd7, 0xff, 0x57, 0xc9, 0x1b, 0xdb, 0x4a, 0xde, 0xdc, 0x28, 0xf9, 0x4d, 0xb5, 0x5b, 0x3b, 0xab,
	0xdd, 0xde, 0x59, 0xed, 0xce, 0x2d, 0x6a, 0xab, 0x1f, 0x52, 0x1b, 0x3e, 0xa0, 0xf6, 0xfe, 0x86,
	0xda, 0x21, 0x68, 0x1b, 0xe3, 0xb4, 0x5a, 0x5c, 0xdf, 0x50, 0xfa, 0x18, 0x9a, 0x8c, 0xf3, 0x88,
	0xe7, 0x13, 0x74, 0xb0, 0x96, 0xd3, 0x16, 0x68, 0x4c, 0x33, 0x2f, 0xf9, 0x12, 0x7a, 0xf2, 0x03,
	0x5f, 0x10, 0xc7, 0xde, 0x8f, 0x4c, 0xaa, 0xaf, 0xd2, 0xae, 0x04, 0x9f, 0x65, 0xd8, 0xe0, 0x8d,
	0x02, 0x07, 0xa7, 0xe9, 0xf5, 0xc6, 0xf0, 0x7e, 0x04, 0x2d, 0xbc, 0xd1, 0x5d, 0xe7, 0x6c, 0xa2,
	0x85, 0x9c, 0x05, 0x8d, 0x95, 0xcc, 0xa9, 0x22, 0x8d, 0xd5, 0xf6, 0xe1, 0xdc, 0x18, 0xad, 0xc6,
	0x0e, 0xa3, 0x35, 0x78, 0x0d, 0xbd, 0x92, 0x84, 0x78, 0xf2, 0x17, 0xd0, 0xe5, 0x2c, 0xf3, 0xbb,
	0x29, 0x5f, 0x48, 0x22, 0x2a, 0xdd, 0x2f, 0xb0, 0x19, 0x5f, 0xdc, 0xa9, 0x0a, 0x7f, 0x29, 0xb8,
	0xc5, 0xbc, 0xd0, 0x67, 0x8b, 0x5d, 0x84, 0xb8, 0xbb, 0x87, 0x6f, 0x6d, 0xea, 0xc1, 0x2f, 0xd8,
	0x07, 0x55, 0x42, 0x42, 0x94, 0x4f, 0x41, 0xf5, 0x25, 0xb6, 0x60, 0x19, 0xa3, 0x0e, 0x2d, 0x81,
	0x3b, 0xd5, 0x23, 0x86, 0x7b, 0xe6, 0x95, 0x97, 0xe4, 0xa6, 0xc3, 0x7e, 0x96, 0x2c, 0xb6, 0xe9,
	0x71, 0x37, 0x93, 0x3f, 0xf8, 0x47, 0x81, 0xa3, 0x4a, 0xd6, 0x7c, 0xe7, 0x9c, 0x05, 0x61, 0x10,
	0x5f, 0xe1, 0xfb, 0xb6, 0x24, 0xc7, 0x11, 0x8c, 0x53, 0xdf, 0xc7, 0x33, 0x32, 0x47, 0x87, 0x16,
	0xa6, 0xf0, 0x5c, 0x7a, 0xc1, 0x22, 0xe5, 0x4c, 0xd6, 0x02, 0x3d, 0xb9, 0x59, 0x21, 0xdc, 0xdc,
	0x4a, 0xb8, 0x75, 0xfb, 0xaa, 0xda, 0x28, 0x77, 0x7b, 0x97, 0x3e, 0x3f, 0x87, 0x07, 0x95, 0x17,
	0x4a, 0x51, 0xcd, 0x75, 0x01, 0xb7, 0x77, 0x5b, 0xd1, 0x22, 0xb5, 0x6a, 0x8b, 0x9c, 0x1c, 0x43,
	0xa7, 0x20, 0x45, 0xda, 0x50, 0xa7, 0xb3, 0x53, 0x6d, 0x4f, 0x7c, 0xcc, 0x26, 0x96, 0xa6, 0x88,
	0x0f, 0x73, 0x7c, 0xa1, 0xd5, 0x4e, 0xde, 0x2a, 0xd0, 0xca, 0x5a, 0x80, 0xb4, 0xa0, 0x36, 0x3e,
	0xc7, 0x20, 0x0d, 0xba, 0x43, 0xe7, 0xb9, 0x31, 0x1a, 0x5a, 0xae, 0x65, 0x4c, 0x0d, 0x8c, 0xee,
	0x81, 0x6a, 0x9d, 0xba, 0x67, 0xc6, 0x70, 0x64, 0x5b, 0x5a, 0x8d, 0x1c, 0x42, 0xcf, 0x18, 0x51,
	0xdb, 0xb0, 0x5e, 0xb8, 0x17, 0xc6, 0x0b, 0x84, 0xea, 0x02, 0xc2, 0x4f, 0xd7, 0x34, 0x1c, 0xd3,
	0x1e, 0x89, 0xa8, 0x06, 0xca, 0x48, 0x0c, 0x67, 0x3c, 0x7d, 0x6a, 0x53, 0x77, 0x7c, 0x61, 0x3b,
	0xee, 0x98, 0x5a, 0x36, 0xd5, 0x9a, 0x22, 0x74, 0xe6, 0x9c, 0x3b, 0xe3, 0x1f, 0x1c, 0xd7, 0xa6,
	0x74, 0x4c, 0xb5, 0xd7, 0xe4, 0x00, 0xf6, 0x87, 0xce, 0x70, 0x5a, 0x64, 0xf8, 0x15, 0x01, 0x10,
	0xd7, 0xe5, 0xf6, 0x6f, 0x0a, 0x6e, 0x17, 0xd5, 0x7c, 0x6a, 0x4c, 0x5d, 0x0b, 0x8f, 0x69, 0xbf,
	0x2b, 0x22, 0xc0, 0x1c, 0x0f, 0x9d, 0x49, 0x06, 0xfc, 0xa1, 0xe0, 0xad, 0x5d, 0x4c, 0x3e, 0x75,
	0xe5, 0x31, 0xdb, 0xd6, 0xde, 0x48, 0x88, 0xda, 0x67, 0x33, 0xc7, 0xca, 0xf3, 0xfc, 0xa9, 0x9c,
	0x3c, 0xc2, 0x87, 0xac, 0xa7, 0x0b, 0xb3, 0x9a, 0xa3, 0xa1, 0x9d, 0x1d, 0x9a, 0xe0, 0xc3, 0x4b,
	0x80, 0xda, 0xe6, 0x73, 0x4d, 0x79, 0xf2, 0xaf, 0x02, 0xfd, 0xbc, 0xd1, 0x26, 0xd9, 0x72, 0x26,
	0xdf, 0x63, 0x4c, 0xb9, 0x81, 0xc9, 0x27, 0x65, 0x07, 0xdc, 0xf8, 0xcd, 0x39, 0x7a, 0xf0, 0x7e,
	0x27, 0x0e, 0xeb, 0x60, 0x8f, 0x7c, 0x07, 0x9d, 0x62, 0xa9, 0x11, 0x7d, 0x1d, 0xf8, 0xce, 0xb2,
	0x3d, 0xba, 0xff, 0x1e, 0x4f, 0x76, 0x5e, 0x10, 0x29, 0x57, 0x40, 0x95, 0xc8, 0x8d, 0x4d, 0x55,
	0x25, 0xf2, 0xce, 0xd6, 0x18, 0xec, 0xbd, 0x6c, 0xc9, 0x7f, 0xb6, 0x6f, 0xff, 0x0b, 0x00, 0x00,
	0xff, 0xff, 0x81, 0x1d, 0xc0, 0xbf, 0xc4, 0x09, 0x00, 0x00,
}
