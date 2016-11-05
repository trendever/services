// Code generated by protoc-gen-go.
// source: payment.proto
// DO NOT EDIT!

/*
Package payment is a generated protocol buffer package.

It is generated from these files:
	payment.proto

It has these top-level messages:
	UsualData
	ChatMessageNewOrder
	ChatMessagePaymentFinished
	ChatMessageOrderCancelled
	OrderData
	CreateOrderRequest
	GetOrderRequest
	GetOrderReply
	UpdateServiceDataRequest
	UpdateServiceDataReply
	PaymentNotification
	CreateOrderReply
	BuyOrderRequest
	BuyOrderReply
	CancelOrderRequest
	CancelOrderReply
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

type Event int32

const (
	Event_Created    Event = 0
	Event_Cancelled  Event = 1
	Event_PayFailed  Event = 2
	Event_PaySuccess Event = 3
)

var Event_name = map[int32]string{
	0: "Created",
	1: "Cancelled",
	2: "PayFailed",
	3: "PaySuccess",
}
var Event_value = map[string]int32{
	"Created":    0,
	"Cancelled":  1,
	"PayFailed":  2,
	"PaySuccess": 3,
}

func (x Event) String() string {
	return proto.EnumName(Event_name, int32(x))
}
func (Event) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

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
func (Currency) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Errors int32

const (
	Errors_OK Errors = 0
	// internal errors
	Errors_INVALID_DATA       Errors = 1
	Errors_DB_FAILED          Errors = 2
	Errors_ALREADY_PAYED      Errors = 3
	Errors_PAY_CANCELLED      Errors = 4
	Errors_ANOTHER_OPEN_ORDER Errors = 5
	Errors_ALREADY_CANCELLED  Errors = 6
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
	Errors_NATS_FAILED  Errors = 133
)

var Errors_name = map[int32]string{
	0:   "OK",
	1:   "INVALID_DATA",
	2:   "DB_FAILED",
	3:   "ALREADY_PAYED",
	4:   "PAY_CANCELLED",
	5:   "ANOTHER_OPEN_ORDER",
	6:   "ALREADY_CANCELLED",
	126: "UNKNOWN_ERROR",
	127: "INIT_FAILED",
	128: "PAY_FAILED",
	129: "CHAT_DOWN",
	130: "COINS_DOWN",
	131: "CANT_PAY_FEE",
	132: "REFUND_ERROR",
	133: "NATS_FAILED",
}
var Errors_value = map[string]int32{
	"OK":                 0,
	"INVALID_DATA":       1,
	"DB_FAILED":          2,
	"ALREADY_PAYED":      3,
	"PAY_CANCELLED":      4,
	"ANOTHER_OPEN_ORDER": 5,
	"ALREADY_CANCELLED":  6,
	"UNKNOWN_ERROR":      126,
	"INIT_FAILED":        127,
	"PAY_FAILED":         128,
	"CHAT_DOWN":          129,
	"COINS_DOWN":         130,
	"CANT_PAY_FEE":       131,
	"REFUND_ERROR":       132,
	"NATS_FAILED":        133,
}

func (x Errors) String() string {
	return proto.EnumName(Errors_name, int32(x))
}
func (Errors) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

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
func (Direction) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

// service_data structs
type UsualData struct {
	Direction      Direction `protobuf:"varint,1,opt,name=direction,enum=payment.Direction" json:"direction,omitempty"`
	ConversationId uint64    `protobuf:"varint,2,opt,name=conversation_id,json=conversationId" json:"conversation_id,omitempty"`
	MessageId      uint64    `protobuf:"varint,3,opt,name=message_id,json=messageId" json:"message_id,omitempty"`
}

func (m *UsualData) Reset()                    { *m = UsualData{} }
func (m *UsualData) String() string            { return proto.CompactTextString(m) }
func (*UsualData) ProtoMessage()               {}
func (*UsualData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// chat messages
type ChatMessageNewOrder struct {
	PayId    uint64   `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Amount   uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,3,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
}

func (m *ChatMessageNewOrder) Reset()                    { *m = ChatMessageNewOrder{} }
func (m *ChatMessageNewOrder) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageNewOrder) ProtoMessage()               {}
func (*ChatMessageNewOrder) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

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
func (*ChatMessagePaymentFinished) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type ChatMessageOrderCancelled struct {
	PayId  uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *ChatMessageOrderCancelled) Reset()                    { *m = ChatMessageOrderCancelled{} }
func (m *ChatMessageOrderCancelled) String() string            { return proto.CompactTextString(m) }
func (*ChatMessageOrderCancelled) ProtoMessage()               {}
func (*ChatMessageOrderCancelled) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type OrderData struct {
	Amount   uint64   `protobuf:"varint,1,opt,name=amount" json:"amount,omitempty"`
	Currency Currency `protobuf:"varint,2,opt,name=currency,enum=payment.Currency" json:"currency,omitempty"`
	Gateway  string   `protobuf:"bytes,3,opt,name=gateway" json:"gateway,omitempty"`
	UserId   uint64   `protobuf:"varint,4,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// payment of our service
	ServiceName string `protobuf:"bytes,5,opt,name=service_name,json=serviceName" json:"service_name,omitempty"`
	ServiceData string `protobuf:"bytes,6,opt,name=service_data,json=serviceData" json:"service_data,omitempty"`
	// p2p payment
	LeadId         uint64 `protobuf:"varint,7,opt,name=lead_id,json=leadId" json:"lead_id,omitempty"`
	ShopCardNumber string `protobuf:"bytes,10,opt,name=shop_card_number,json=shopCardNumber" json:"shop_card_number,omitempty"`
	// in trendcoins
	CommissionFee uint64 `protobuf:"varint,11,opt,name=commission_fee,json=commissionFee" json:"commission_fee,omitempty"`
	// user id, usually supplier
	CommissionSource uint64 `protobuf:"varint,12,opt,name=commission_source,json=commissionSource" json:"commission_source,omitempty"`
	Cancelled        bool   `protobuf:"varint,13,opt,name=cancelled" json:"cancelled,omitempty"`
}

func (m *OrderData) Reset()                    { *m = OrderData{} }
func (m *OrderData) String() string            { return proto.CompactTextString(m) }
func (*OrderData) ProtoMessage()               {}
func (*OrderData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CreateOrderRequest struct {
	Data *OrderData `protobuf:"bytes,1,opt,name=data" json:"data,omitempty"`
}

func (m *CreateOrderRequest) Reset()                    { *m = CreateOrderRequest{} }
func (m *CreateOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderRequest) ProtoMessage()               {}
func (*CreateOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *CreateOrderRequest) GetData() *OrderData {
	if m != nil {
		return m.Data
	}
	return nil
}

type GetOrderRequest struct {
	Id uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetOrderRequest) Reset()                    { *m = GetOrderRequest{} }
func (m *GetOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*GetOrderRequest) ProtoMessage()               {}
func (*GetOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type GetOrderReply struct {
	Order *OrderData `protobuf:"bytes,1,opt,name=order" json:"order,omitempty"`
}

func (m *GetOrderReply) Reset()                    { *m = GetOrderReply{} }
func (m *GetOrderReply) String() string            { return proto.CompactTextString(m) }
func (*GetOrderReply) ProtoMessage()               {}
func (*GetOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *GetOrderReply) GetOrder() *OrderData {
	if m != nil {
		return m.Order
	}
	return nil
}

type UpdateServiceDataRequest struct {
	Id      uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	NewData string `protobuf:"bytes,2,opt,name=new_data,json=newData" json:"new_data,omitempty"`
}

func (m *UpdateServiceDataRequest) Reset()                    { *m = UpdateServiceDataRequest{} }
func (m *UpdateServiceDataRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateServiceDataRequest) ProtoMessage()               {}
func (*UpdateServiceDataRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type UpdateServiceDataReply struct {
}

func (m *UpdateServiceDataReply) Reset()                    { *m = UpdateServiceDataReply{} }
func (m *UpdateServiceDataReply) String() string            { return proto.CompactTextString(m) }
func (*UpdateServiceDataReply) ProtoMessage()               {}
func (*UpdateServiceDataReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type PaymentNotification struct {
	Id            uint64     `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Data          *OrderData `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
	Event         Event      `protobuf:"varint,3,opt,name=event,enum=payment.Event" json:"event,omitempty"`
	InvokerUserId uint64     `protobuf:"varint,4,opt,name=invoker_user_id,json=invokerUserId" json:"invoker_user_id,omitempty"`
}

func (m *PaymentNotification) Reset()                    { *m = PaymentNotification{} }
func (m *PaymentNotification) String() string            { return proto.CompactTextString(m) }
func (*PaymentNotification) ProtoMessage()               {}
func (*PaymentNotification) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *PaymentNotification) GetData() *OrderData {
	if m != nil {
		return m.Data
	}
	return nil
}

type CreateOrderReply struct {
	Id           uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CreateOrderReply) Reset()                    { *m = CreateOrderReply{} }
func (m *CreateOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CreateOrderReply) ProtoMessage()               {}
func (*CreateOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

type BuyOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	Ip    string `protobuf:"bytes,2,opt,name=ip" json:"ip,omitempty"`
}

func (m *BuyOrderRequest) Reset()                    { *m = BuyOrderRequest{} }
func (m *BuyOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderRequest) ProtoMessage()               {}
func (*BuyOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type BuyOrderReply struct {
	RedirectUrl  string `protobuf:"bytes,1,opt,name=redirect_url,json=redirectUrl" json:"redirect_url,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *BuyOrderReply) Reset()                    { *m = BuyOrderReply{} }
func (m *BuyOrderReply) String() string            { return proto.CompactTextString(m) }
func (*BuyOrderReply) ProtoMessage()               {}
func (*BuyOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

type CancelOrderRequest struct {
	PayId uint64 `protobuf:"varint,1,opt,name=pay_id,json=payId" json:"pay_id,omitempty"`
	// userID just to log it
	UserId uint64 `protobuf:"varint,5,opt,name=user_id,json=userId" json:"user_id,omitempty"`
}

func (m *CancelOrderRequest) Reset()                    { *m = CancelOrderRequest{} }
func (m *CancelOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderRequest) ProtoMessage()               {}
func (*CancelOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

type CancelOrderReply struct {
	Cancelled    bool   `protobuf:"varint,1,opt,name=cancelled" json:"cancelled,omitempty"`
	Error        Errors `protobuf:"varint,2,opt,name=error,enum=payment.Errors" json:"error,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage" json:"error_message,omitempty"`
}

func (m *CancelOrderReply) Reset()                    { *m = CancelOrderReply{} }
func (m *CancelOrderReply) String() string            { return proto.CompactTextString(m) }
func (*CancelOrderReply) ProtoMessage()               {}
func (*CancelOrderReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func init() {
	proto.RegisterType((*UsualData)(nil), "payment.UsualData")
	proto.RegisterType((*ChatMessageNewOrder)(nil), "payment.ChatMessageNewOrder")
	proto.RegisterType((*ChatMessagePaymentFinished)(nil), "payment.ChatMessagePaymentFinished")
	proto.RegisterType((*ChatMessageOrderCancelled)(nil), "payment.ChatMessageOrderCancelled")
	proto.RegisterType((*OrderData)(nil), "payment.OrderData")
	proto.RegisterType((*CreateOrderRequest)(nil), "payment.CreateOrderRequest")
	proto.RegisterType((*GetOrderRequest)(nil), "payment.GetOrderRequest")
	proto.RegisterType((*GetOrderReply)(nil), "payment.GetOrderReply")
	proto.RegisterType((*UpdateServiceDataRequest)(nil), "payment.UpdateServiceDataRequest")
	proto.RegisterType((*UpdateServiceDataReply)(nil), "payment.UpdateServiceDataReply")
	proto.RegisterType((*PaymentNotification)(nil), "payment.PaymentNotification")
	proto.RegisterType((*CreateOrderReply)(nil), "payment.CreateOrderReply")
	proto.RegisterType((*BuyOrderRequest)(nil), "payment.BuyOrderRequest")
	proto.RegisterType((*BuyOrderReply)(nil), "payment.BuyOrderReply")
	proto.RegisterType((*CancelOrderRequest)(nil), "payment.CancelOrderRequest")
	proto.RegisterType((*CancelOrderReply)(nil), "payment.CancelOrderReply")
	proto.RegisterEnum("payment.Event", Event_name, Event_value)
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
	GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*GetOrderReply, error)
	UpdateServiceData(ctx context.Context, in *UpdateServiceDataRequest, opts ...grpc.CallOption) (*UpdateServiceDataReply, error)
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

func (c *paymentServiceClient) GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*GetOrderReply, error) {
	out := new(GetOrderReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/GetOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *paymentServiceClient) UpdateServiceData(ctx context.Context, in *UpdateServiceDataRequest, opts ...grpc.CallOption) (*UpdateServiceDataReply, error) {
	out := new(UpdateServiceDataReply)
	err := grpc.Invoke(ctx, "/payment.PaymentService/UpdateServiceData", in, out, c.cc, opts...)
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
	GetOrder(context.Context, *GetOrderRequest) (*GetOrderReply, error)
	UpdateServiceData(context.Context, *UpdateServiceDataRequest) (*UpdateServiceDataReply, error)
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

func _PaymentService_GetOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).GetOrder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/GetOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).GetOrder(ctx, req.(*GetOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PaymentService_UpdateServiceData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateServiceDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).UpdateServiceData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/UpdateServiceData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).UpdateServiceData(ctx, req.(*UpdateServiceDataRequest))
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
		{
			MethodName: "GetOrder",
			Handler:    _PaymentService_GetOrder_Handler,
		},
		{
			MethodName: "UpdateServiceData",
			Handler:    _PaymentService_UpdateServiceData_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("payment.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1097 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x56, 0x5d, 0x6f, 0xe3, 0x44,
	0x17, 0x5e, 0xa7, 0x4d, 0xd2, 0x9c, 0x34, 0x89, 0x3b, 0xab, 0xb7, 0xaf, 0x5b, 0x40, 0x50, 0x43,
	0x61, 0x55, 0xb4, 0x2b, 0x54, 0x6e, 0x40, 0x42, 0x48, 0xae, 0xed, 0xee, 0x46, 0x2d, 0x4e, 0x35,
	0x49, 0x16, 0xf5, 0xca, 0xf2, 0x3a, 0xd3, 0xad, 0x45, 0x62, 0x07, 0x3b, 0x6e, 0x55, 0x09, 0x2d,
	0xcb, 0xc7, 0xfe, 0x03, 0xee, 0xf8, 0x73, 0xfc, 0x11, 0x24, 0xce, 0xcc, 0xd8, 0xb1, 0x93, 0x36,
	0xdd, 0xbd, 0xd8, 0xbb, 0x39, 0xcf, 0x9c, 0x8f, 0x67, 0xce, 0x97, 0x0d, 0xad, 0xa9, 0x77, 0x33,
	0x61, 0xe1, 0xec, 0xc9, 0x34, 0x8e, 0x66, 0x11, 0xa9, 0x67, 0xa2, 0xfe, 0x46, 0x81, 0xc6, 0x30,
	0x49, 0xbd, 0xb1, 0xe5, 0xcd, 0x3c, 0xf2, 0x15, 0x34, 0x46, 0x41, 0xcc, 0xfc, 0x59, 0x10, 0x85,
	0x9a, 0xf2, 0x89, 0xf2, 0xa8, 0x7d, 0x48, 0x9e, 0xe4, 0x96, 0x56, 0x7e, 0x43, 0x0b, 0x25, 0xf2,
	0x05, 0x74, 0xfc, 0x28, 0xbc, 0x62, 0x71, 0xe2, 0x71, 0xd9, 0x0d, 0x46, 0x5a, 0x05, 0xed, 0xd6,
	0x69, 0xbb, 0x0c, 0x77, 0x47, 0xe4, 0x23, 0x80, 0x09, 0x4b, 0x12, 0xef, 0x25, 0xe3, 0x3a, 0x6b,
	0x42, 0xa7, 0x91, 0x21, 0xdd, 0x91, 0x9e, 0xc0, 0x43, 0xf3, 0xd2, 0x9b, 0xfd, 0x20, 0x01, 0x87,
	0x5d, 0xf7, 0xe2, 0x11, 0x8b, 0xc9, 0xff, 0xa0, 0x86, 0xe1, 0xb9, 0x85, 0x22, 0x2c, 0xaa, 0x28,
	0xa1, 0xb3, 0x6d, 0xa8, 0x79, 0x93, 0x28, 0x0d, 0x67, 0x59, 0xb0, 0x4c, 0x22, 0x8f, 0x61, 0xc3,
	0x4f, 0xe3, 0x98, 0x85, 0xfe, 0x8d, 0x08, 0xd1, 0x3e, 0xdc, 0x9a, 0xd3, 0x37, 0xb3, 0x0b, 0x3a,
	0x57, 0xd1, 0xff, 0x51, 0x60, 0xb7, 0x14, 0xf5, 0x4c, 0x6a, 0x1e, 0x07, 0x61, 0x90, 0x5c, 0xb2,
	0xd1, 0xaa, 0xe0, 0x1a, 0xd4, 0x93, 0xd4, 0xf7, 0xd1, 0x46, 0xc4, 0xd8, 0xa0, 0xb9, 0xc8, 0x6f,
	0x2e, 0xbc, 0x60, 0x9c, 0xc6, 0x4c, 0x5b, 0x97, 0x37, 0x99, 0x58, 0x22, 0x5c, 0x5d, 0x49, 0xb8,
	0xf6, 0x56, 0xc2, 0x8b, 0xf5, 0xa9, 0xbf, 0x43, 0x7d, 0xf4, 0x13, 0xd8, 0x29, 0xbd, 0x50, 0x24,
	0xd5, 0xf4, 0x42, 0x9f, 0x8d, 0xc7, 0xab, 0x1f, 0xf8, 0x7f, 0xa8, 0xa7, 0x09, 0x8b, 0x8b, 0x5a,
	0xd6, 0xb8, 0x88, 0x45, 0x7a, 0xbd, 0x06, 0x0d, 0xe1, 0x42, 0x34, 0x4b, 0xf1, 0x26, 0x65, 0xe5,
	0x9b, 0x2a, 0x6f, 0x7f, 0x13, 0x26, 0xed, 0xa5, 0x37, 0x63, 0xd7, 0x9e, 0x2c, 0x59, 0x83, 0xe6,
	0x62, 0x99, 0xc7, 0x7a, 0x99, 0x07, 0xd9, 0x83, 0x4d, 0x3c, 0x5c, 0x05, 0x3e, 0x73, 0x43, 0x6f,
	0xc2, 0x44, 0x4e, 0x1b, 0xb4, 0x99, 0x61, 0x0e, 0x42, 0x65, 0x95, 0x11, 0x92, 0x15, 0xc9, 0x2d,
	0x54, 0x04, 0x7f, 0x74, 0x3f, 0x66, 0xde, 0x88, 0xbb, 0xaf, 0x4b, 0xf7, 0x5c, 0x44, 0xf7, 0x8f,
	0x40, 0x4d, 0x2e, 0xa3, 0xa9, 0xeb, 0x7b, 0xf1, 0xc8, 0x0d, 0xd3, 0xc9, 0x0b, 0x16, 0x6b, 0x20,
	0xec, 0xdb, 0x1c, 0x37, 0x11, 0x76, 0x04, 0x4a, 0xf6, 0x01, 0xdb, 0x7c, 0x32, 0x09, 0x92, 0x84,
	0xf7, 0xfe, 0x05, 0x63, 0x5a, 0x53, 0x78, 0x6a, 0x15, 0xe8, 0x31, 0x63, 0xe4, 0x4b, 0xd8, 0x2a,
	0xa9, 0x25, 0x51, 0x1a, 0xfb, 0x4c, 0xdb, 0x14, 0x9a, 0x6a, 0x71, 0xd1, 0x17, 0x38, 0xf9, 0x10,
	0x1a, 0x7e, 0x5e, 0x21, 0xad, 0x25, 0xda, 0xa8, 0x00, 0xf4, 0xef, 0x80, 0x98, 0x31, 0xc3, 0x04,
	0x89, 0x3a, 0x50, 0xf6, 0x73, 0xca, 0x92, 0x19, 0xf9, 0x1c, 0xd6, 0xc5, 0x2b, 0x79, 0x21, 0x9a,
	0xa5, 0x96, 0x98, 0x17, 0x8b, 0x8a, 0x7b, 0x7d, 0x0f, 0x3a, 0x4f, 0xd9, 0x6c, 0xc1, 0xb4, 0x0d,
	0x95, 0x79, 0xfd, 0xf1, 0xa4, 0x7f, 0x0b, 0xad, 0x42, 0x65, 0x3a, 0xbe, 0xc1, 0x6c, 0x54, 0x23,
	0x2e, 0xdd, 0xe3, 0x5c, 0x2a, 0xe8, 0x36, 0x68, 0xc3, 0x29, 0xc6, 0x61, 0xfd, 0x22, 0xcb, 0x2b,
	0xc2, 0x90, 0x1d, 0xd8, 0x08, 0xd9, 0xb5, 0xac, 0x4d, 0x45, 0x96, 0x1d, 0x65, 0x6e, 0xa1, 0x6b,
	0xb0, 0x7d, 0x87, 0x1b, 0xa4, 0xa2, 0xff, 0xad, 0xc0, 0xc3, 0x6c, 0x48, 0x9d, 0x68, 0x16, 0x5c,
	0x04, 0xbe, 0xd8, 0x2e, 0xb7, 0x9c, 0xe7, 0xe9, 0xa8, 0xdc, 0x9f, 0x0e, 0xf2, 0x19, 0x54, 0xd9,
	0x15, 0x5e, 0x64, 0xbb, 0xa2, 0x3d, 0x57, 0xb4, 0x39, 0x4a, 0xe5, 0x25, 0x7a, 0xeb, 0x04, 0xe1,
	0x55, 0xf4, 0x13, 0x76, 0xe2, 0x62, 0x3b, 0xb6, 0x32, 0x78, 0x28, 0xa7, 0x23, 0x04, 0x75, 0xa1,
	0x34, 0x3c, 0x79, 0xcb, 0xcc, 0xf6, 0x31, 0x62, 0x1c, 0x47, 0x71, 0x36, 0x18, 0x9d, 0x22, 0x22,
	0x47, 0x13, 0x2a, 0x6f, 0xc9, 0xa7, 0xd0, 0x12, 0x07, 0x37, 0x5b, 0x90, 0xd9, 0x64, 0x6c, 0x0a,
	0x30, 0x9b, 0x65, 0xfd, 0x1b, 0xe8, 0x1c, 0xa5, 0x37, 0x0b, 0xc5, 0x5c, 0x31, 0xd0, 0x9c, 0xc5,
	0x34, 0x4b, 0x33, 0x9e, 0xf4, 0x57, 0xd0, 0x2a, 0x2c, 0x39, 0x4d, 0x9c, 0x96, 0x98, 0xc9, 0xa5,
	0xe1, 0xa6, 0xf1, 0x58, 0x58, 0xe3, 0xb4, 0xe4, 0xd8, 0x30, 0x1e, 0xbf, 0x57, 0xe6, 0x16, 0x36,
	0xb1, 0xe8, 0xe8, 0x77, 0x21, 0x5f, 0xda, 0x02, 0xd5, 0x85, 0x6d, 0xf4, 0x0b, 0xe6, 0xbb, 0xec,
	0x85, 0x3f, 0x64, 0x61, 0x78, 0x94, 0xa5, 0xe1, 0x79, 0x9f, 0x6f, 0x38, 0x38, 0x82, 0xaa, 0xe8,
	0x12, 0xd2, 0x84, 0xba, 0x2c, 0xfb, 0x48, 0x7d, 0x40, 0x5a, 0xd0, 0x98, 0xaf, 0x57, 0x55, 0xe1,
	0x22, 0xf6, 0xeb, 0x31, 0x7e, 0x04, 0x50, 0xac, 0x60, 0x1d, 0x00, 0xc5, 0xbe, 0xfc, 0x5a, 0xa8,
	0x6b, 0x07, 0xfb, 0xb0, 0x91, 0x2f, 0x44, 0x52, 0x87, 0x35, 0x3a, 0x3c, 0x42, 0x17, 0x78, 0x18,
	0xf6, 0x2d, 0x34, 0xc6, 0x83, 0xd9, 0x3b, 0x53, 0x2b, 0x07, 0x7f, 0x55, 0xa0, 0x26, 0x19, 0x92,
	0x1a, 0x54, 0x7a, 0x27, 0xa8, 0xa4, 0xc2, 0x66, 0xd7, 0x79, 0x6e, 0x9c, 0x76, 0x2d, 0xd7, 0x32,
	0x06, 0x86, 0x0c, 0x65, 0x1d, 0xb9, 0xc7, 0x46, 0xf7, 0xd4, 0xb6, 0x30, 0xd4, 0x16, 0xb4, 0x8c,
	0x53, 0x6a, 0x1b, 0xd6, 0xb9, 0x7b, 0x66, 0x9c, 0x23, 0xb4, 0xc6, 0x21, 0x3c, 0xba, 0xa6, 0xe1,
	0x98, 0xf6, 0x29, 0xd7, 0x5a, 0xc7, 0x15, 0x4e, 0x0c, 0xa7, 0x37, 0x78, 0x66, 0x53, 0xb7, 0x77,
	0x66, 0x3b, 0x6e, 0x8f, 0x5a, 0x36, 0x55, 0xab, 0x58, 0x8a, 0xad, 0xdc, 0xba, 0x50, 0xaf, 0x71,
	0x0f, 0x43, 0xe7, 0xc4, 0xe9, 0xfd, 0xe8, 0xb8, 0x36, 0xa5, 0x3d, 0xaa, 0xbe, 0x22, 0x1d, 0x68,
	0x76, 0x9d, 0xee, 0x20, 0x0f, 0xfc, 0x2b, 0x02, 0xc0, 0xa3, 0x64, 0xf2, 0x6b, 0x05, 0x1f, 0xdd,
	0x30, 0x9f, 0x19, 0x03, 0xd7, 0x42, 0x33, 0xf5, 0x37, 0x85, 0x2b, 0x98, 0xbd, 0xae, 0xd3, 0x97,
	0xc0, 0xef, 0x0a, 0x7a, 0xdd, 0xc4, 0x20, 0x03, 0x57, 0x98, 0xd9, 0xb6, 0xfa, 0x87, 0x80, 0xa8,
	0x7d, 0x3c, 0x74, 0xac, 0x2c, 0xce, 0x9f, 0x0a, 0xbe, 0xb8, 0xe9, 0x18, 0x83, 0x7e, 0xee, 0xf8,
	0x8d, 0x72, 0xf0, 0x18, 0x5f, 0x3c, 0xff, 0x0f, 0x41, 0x1e, 0xe6, 0x69, 0xd7, 0x96, 0x6e, 0xfa,
	0x98, 0xa1, 0x02, 0xa0, 0xb6, 0xf9, 0x5c, 0x55, 0x0e, 0xff, 0xad, 0x40, 0x3b, 0x5b, 0x1e, 0xd9,
	0x62, 0x21, 0x4f, 0x51, 0xa7, 0x98, 0x58, 0xf2, 0x41, 0xf1, 0x99, 0xba, 0xb5, 0x62, 0x77, 0x77,
	0xee, 0xbe, 0xe4, 0x6b, 0xe9, 0x01, 0xf9, 0x1e, 0x36, 0xf2, 0x81, 0x22, 0xda, 0x5c, 0x71, 0x69,
	0x3a, 0x77, 0xb7, 0xef, 0xb8, 0x91, 0xf6, 0x9c, 0x48, 0xd1, 0xca, 0x65, 0x22, 0xb7, 0xc6, 0xa4,
	0x4c, 0x64, 0xa9, 0xfb, 0x25, 0x91, 0x7c, 0x7b, 0x97, 0x88, 0x2c, 0xed, 0xfc, 0x12, 0x91, 0x85,
	0x55, 0x8f, 0xf6, 0xe7, 0xb0, 0x75, 0x6b, 0xf7, 0x92, 0xbd, 0xb9, 0xfa, 0xaa, 0xf5, 0xbe, 0xfb,
	0xf1, 0x7d, 0x2a, 0xc2, 0xf5, 0x8b, 0x9a, 0xf8, 0xf3, 0xfc, 0xfa, 0xbf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x1c, 0x45, 0x9b, 0x8a, 0x8a, 0x0a, 0x00, 0x00,
}
