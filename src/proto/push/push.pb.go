// Code generated by protoc-gen-gogo.
// source: push.proto
// DO NOT EDIT!

/*
	Package push is a generated protocol buffer package.

	It is generated from these files:
		push.proto

	It has these top-level messages:
		Receiver
		PushMessage
		PushRequest
		PushResult
*/
package push

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ServiceType int32

const (
	// Apple Push Notifications
	ServiceType_APN ServiceType = 0
	// Firebase Cloud Messaging
	ServiceType_FCM ServiceType = 1
)

var ServiceType_name = map[int32]string{
	0: "APN",
	1: "FCM",
}
var ServiceType_value = map[string]int32{
	"APN": 0,
	"FCM": 1,
}

func (x ServiceType) String() string {
	return proto.EnumName(ServiceType_name, int32(x))
}
func (ServiceType) EnumDescriptor() ([]byte, []int) { return fileDescriptorPush, []int{0} }

type Priority int32

const (
	Priority_NORMAL Priority = 0
	Priority_HING   Priority = 1
)

var Priority_name = map[int32]string{
	0: "NORMAL",
	1: "HING",
}
var Priority_value = map[string]int32{
	"NORMAL": 0,
	"HING":   1,
}

func (x Priority) String() string {
	return proto.EnumName(Priority_name, int32(x))
}
func (Priority) EnumDescriptor() ([]byte, []int) { return fileDescriptorPush, []int{1} }

type Receiver struct {
	Service ServiceType `protobuf:"varint,1,opt,name=service,proto3,enum=push.ServiceType" json:"service,omitempty"`
	Token   string      `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (m *Receiver) Reset()                    { *m = Receiver{} }
func (m *Receiver) String() string            { return proto.CompactTextString(m) }
func (*Receiver) ProtoMessage()               {}
func (*Receiver) Descriptor() ([]byte, []int) { return fileDescriptorPush, []int{0} }

type PushMessage struct {
	Priority Priority `protobuf:"varint,1,opt,name=priority,proto3,enum=push.Priority" json:"priority,omitempty"`
	// seconds
	TimeToLive uint64 `protobuf:"varint,2,opt,name=time_to_live,json=timeToLive,proto3" json:"time_to_live,omitempty"`
	// at last one of two fields below should be non-empty
	// valid json data
	Data string `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	// notification body
	Body string `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`
	// will be ignored if body field is empty
	Title string `protobuf:"bytes,5,opt,name=title,proto3" json:"title,omitempty"`
}

func (m *PushMessage) Reset()                    { *m = PushMessage{} }
func (m *PushMessage) String() string            { return proto.CompactTextString(m) }
func (*PushMessage) ProtoMessage()               {}
func (*PushMessage) Descriptor() ([]byte, []int) { return fileDescriptorPush, []int{1} }

type PushRequest struct {
	Receivers []*Receiver  `protobuf:"bytes,1,rep,name=receivers" json:"receivers,omitempty"`
	Message   *PushMessage `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *PushRequest) Reset()                    { *m = PushRequest{} }
func (m *PushRequest) String() string            { return proto.CompactTextString(m) }
func (*PushRequest) ProtoMessage()               {}
func (*PushRequest) Descriptor() ([]byte, []int) { return fileDescriptorPush, []int{2} }

func (m *PushRequest) GetReceivers() []*Receiver {
	if m != nil {
		return m.Receivers
	}
	return nil
}

func (m *PushRequest) GetMessage() *PushMessage {
	if m != nil {
		return m.Message
	}
	return nil
}

type PushResult struct {
}

func (m *PushResult) Reset()                    { *m = PushResult{} }
func (m *PushResult) String() string            { return proto.CompactTextString(m) }
func (*PushResult) ProtoMessage()               {}
func (*PushResult) Descriptor() ([]byte, []int) { return fileDescriptorPush, []int{3} }

func init() {
	proto.RegisterType((*Receiver)(nil), "push.Receiver")
	proto.RegisterType((*PushMessage)(nil), "push.PushMessage")
	proto.RegisterType((*PushRequest)(nil), "push.PushRequest")
	proto.RegisterType((*PushResult)(nil), "push.PushResult")
	proto.RegisterEnum("push.ServiceType", ServiceType_name, ServiceType_value)
	proto.RegisterEnum("push.Priority", Priority_name, Priority_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for PushService service

type PushServiceClient interface {
	Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushResult, error)
}

type pushServiceClient struct {
	cc *grpc.ClientConn
}

func NewPushServiceClient(cc *grpc.ClientConn) PushServiceClient {
	return &pushServiceClient{cc}
}

func (c *pushServiceClient) Push(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushResult, error) {
	out := new(PushResult)
	err := grpc.Invoke(ctx, "/push.PushService/Push", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for PushService service

type PushServiceServer interface {
	Push(context.Context, *PushRequest) (*PushResult, error)
}

func RegisterPushServiceServer(s *grpc.Server, srv PushServiceServer) {
	s.RegisterService(&_PushService_serviceDesc, srv)
}

func _PushService_Push_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PushServiceServer).Push(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/push.PushService/Push",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PushServiceServer).Push(ctx, req.(*PushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PushService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "push.PushService",
	HandlerType: (*PushServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Push",
			Handler:    _PushService_Push_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "push.proto",
}

func (m *Receiver) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Receiver) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Service != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintPush(dAtA, i, uint64(m.Service))
	}
	if len(m.Token) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintPush(dAtA, i, uint64(len(m.Token)))
		i += copy(dAtA[i:], m.Token)
	}
	return i, nil
}

func (m *PushMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PushMessage) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Priority != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintPush(dAtA, i, uint64(m.Priority))
	}
	if m.TimeToLive != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintPush(dAtA, i, uint64(m.TimeToLive))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintPush(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if len(m.Body) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintPush(dAtA, i, uint64(len(m.Body)))
		i += copy(dAtA[i:], m.Body)
	}
	if len(m.Title) > 0 {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintPush(dAtA, i, uint64(len(m.Title)))
		i += copy(dAtA[i:], m.Title)
	}
	return i, nil
}

func (m *PushRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PushRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Receivers) > 0 {
		for _, msg := range m.Receivers {
			dAtA[i] = 0xa
			i++
			i = encodeVarintPush(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.Message != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintPush(dAtA, i, uint64(m.Message.Size()))
		n1, err := m.Message.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	return i, nil
}

func (m *PushResult) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PushResult) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func encodeFixed64Push(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Push(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintPush(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Receiver) Size() (n int) {
	var l int
	_ = l
	if m.Service != 0 {
		n += 1 + sovPush(uint64(m.Service))
	}
	l = len(m.Token)
	if l > 0 {
		n += 1 + l + sovPush(uint64(l))
	}
	return n
}

func (m *PushMessage) Size() (n int) {
	var l int
	_ = l
	if m.Priority != 0 {
		n += 1 + sovPush(uint64(m.Priority))
	}
	if m.TimeToLive != 0 {
		n += 1 + sovPush(uint64(m.TimeToLive))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovPush(uint64(l))
	}
	l = len(m.Body)
	if l > 0 {
		n += 1 + l + sovPush(uint64(l))
	}
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovPush(uint64(l))
	}
	return n
}

func (m *PushRequest) Size() (n int) {
	var l int
	_ = l
	if len(m.Receivers) > 0 {
		for _, e := range m.Receivers {
			l = e.Size()
			n += 1 + l + sovPush(uint64(l))
		}
	}
	if m.Message != nil {
		l = m.Message.Size()
		n += 1 + l + sovPush(uint64(l))
	}
	return n
}

func (m *PushResult) Size() (n int) {
	var l int
	_ = l
	return n
}

func sovPush(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozPush(x uint64) (n int) {
	return sovPush(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Receiver) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPush
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Receiver: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Receiver: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Service", wireType)
			}
			m.Service = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Service |= (ServiceType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Token = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPush(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPush
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PushMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPush
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PushMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PushMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Priority", wireType)
			}
			m.Priority = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Priority |= (Priority(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeToLive", wireType)
			}
			m.TimeToLive = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeToLive |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Body", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Body = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPush(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPush
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PushRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPush
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PushRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PushRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Receivers", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Receivers = append(m.Receivers, &Receiver{})
			if err := m.Receivers[len(m.Receivers)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPush
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPush
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Message == nil {
				m.Message = &PushMessage{}
			}
			if err := m.Message.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPush(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPush
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PushResult) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPush
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PushResult: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PushResult: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipPush(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPush
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipPush(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPush
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPush
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPush
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthPush
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowPush
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipPush(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthPush = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPush   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("push.proto", fileDescriptorPush) }

var fileDescriptorPush = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x4c, 0x52, 0x4d, 0x6f, 0xe2, 0x30,
	0x14, 0x8c, 0x97, 0x00, 0xe1, 0x05, 0xa1, 0xac, 0xb5, 0x87, 0x68, 0x0f, 0xd9, 0x28, 0x27, 0xc4,
	0xb6, 0x1c, 0xe8, 0xb5, 0x17, 0x5a, 0xa9, 0x1f, 0x12, 0xa1, 0xc8, 0xe5, 0x8e, 0xf8, 0x78, 0x2a,
	0x56, 0xa1, 0x49, 0x63, 0x27, 0x12, 0xff, 0xa4, 0xfd, 0x47, 0x3d, 0xf6, 0x27, 0x54, 0xf4, 0x8f,
	0x54, 0xb6, 0x93, 0x36, 0xb7, 0xf7, 0xc6, 0xa3, 0x79, 0x33, 0x23, 0x03, 0xa4, 0xb9, 0xd8, 0x0e,
	0xd3, 0x2c, 0x91, 0x09, 0xb5, 0xd5, 0x1c, 0xc5, 0xe0, 0x30, 0x5c, 0x23, 0x2f, 0x30, 0xa3, 0xff,
	0xa1, 0x2d, 0x30, 0x2b, 0xf8, 0x1a, 0x7d, 0x12, 0x92, 0x7e, 0x6f, 0xf4, 0x7b, 0xa8, 0xf9, 0xf7,
	0x06, 0x9c, 0x1f, 0x52, 0x64, 0x15, 0x83, 0xfe, 0x81, 0xa6, 0x4c, 0x1e, 0xf1, 0xc9, 0xff, 0x15,
	0x92, 0x7e, 0x87, 0x99, 0x25, 0x7a, 0x25, 0xe0, 0xce, 0x72, 0xb1, 0x8d, 0x51, 0x88, 0xe5, 0x03,
	0xd2, 0x01, 0x38, 0x69, 0xc6, 0x93, 0x8c, 0xcb, 0x43, 0xa9, 0xd9, 0x33, 0x9a, 0xb3, 0x12, 0x65,
	0xdf, 0xef, 0x34, 0x84, 0xae, 0xe4, 0x7b, 0x5c, 0xc8, 0x64, 0xb1, 0xe3, 0x05, 0x6a, 0x61, 0x9b,
	0x81, 0xc2, 0xe6, 0xc9, 0x84, 0x17, 0x48, 0x29, 0xd8, 0x9b, 0xa5, 0x5c, 0xfa, 0x0d, 0x7d, 0x52,
	0xcf, 0x0a, 0x5b, 0x25, 0x9b, 0x83, 0x6f, 0x1b, 0x4c, 0xcd, 0xda, 0x1b, 0x97, 0x3b, 0xf4, 0x9b,
	0xa5, 0x37, 0xb5, 0x44, 0x5b, 0x63, 0x8d, 0xe1, 0x73, 0x8e, 0x42, 0xd2, 0x13, 0xe8, 0x64, 0x65,
	0x72, 0xe1, 0x93, 0xb0, 0xd1, 0x77, 0x2b, 0x6f, 0x55, 0x21, 0xec, 0x87, 0xa0, 0xba, 0xd9, 0x9b,
	0x4c, 0xda, 0x97, 0x5b, 0x75, 0x53, 0x0b, 0xcb, 0x2a, 0x46, 0xd4, 0x05, 0x30, 0x97, 0x44, 0xbe,
	0x93, 0x83, 0x7f, 0xe0, 0xd6, 0x1a, 0xa4, 0x6d, 0x68, 0x8c, 0x67, 0x53, 0xcf, 0x52, 0xc3, 0xd5,
	0x65, 0xec, 0x91, 0x41, 0x08, 0x4e, 0x55, 0x07, 0x05, 0x68, 0x4d, 0xef, 0x58, 0x3c, 0x9e, 0x78,
	0x16, 0x75, 0xc0, 0xbe, 0xb9, 0x9d, 0x5e, 0x7b, 0x64, 0x74, 0x6e, 0xac, 0x97, 0x32, 0xf4, 0x14,
	0x6c, 0xb5, 0xd2, 0x9a, 0x87, 0x32, 0xd5, 0x5f, 0xaf, 0x0e, 0xa9, 0xf3, 0x91, 0x75, 0xe1, 0xbd,
	0x1d, 0x03, 0xf2, 0x7e, 0x0c, 0xc8, 0xc7, 0x31, 0x20, 0x2f, 0x9f, 0x81, 0xb5, 0x6a, 0xe9, 0x2f,
	0x70, 0xf6, 0x15, 0x00, 0x00, 0xff, 0xff, 0xc9, 0x7f, 0xe8, 0xf2, 0x10, 0x02, 0x00, 0x00,
}
