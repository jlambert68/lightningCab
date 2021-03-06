// Code generated by protoc-gen-go. DO NOT EDIT.
// source: customer_ui_api.proto

package customer_ui_api

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

type PriceUnit int32

const (
	PriceUnit_SatoshiPerSecond PriceUnit = 0
)

var PriceUnit_name = map[int32]string{
	0: "SatoshiPerSecond",
}
var PriceUnit_value = map[string]int32{
	"SatoshiPerSecond": 0,
}

func (x PriceUnit) String() string {
	return proto.EnumName(PriceUnit_name, int32(x))
}
func (PriceUnit) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{0}
}

type TimeUnit int32

const (
	TimeUnit_SecondsBetweenPaymentmentRequests      TimeUnit = 0
	TimeUnit_MilliSecondsBetweenPaymentmentRequests TimeUnit = 1
)

var TimeUnit_name = map[int32]string{
	0: "SecondsBetweenPaymentmentRequests",
	1: "MilliSecondsBetweenPaymentmentRequests",
}
var TimeUnit_value = map[string]int32{
	"SecondsBetweenPaymentmentRequests":      0,
	"MilliSecondsBetweenPaymentmentRequests": 1,
}

func (x TimeUnit) String() string {
	return proto.EnumName(TimeUnit_name, int32(x))
}
func (TimeUnit) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{1}
}

// Parameter used for Empty inputs
type EmptyParameter struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EmptyParameter) Reset()         { *m = EmptyParameter{} }
func (m *EmptyParameter) String() string { return proto.CompactTextString(m) }
func (*EmptyParameter) ProtoMessage()    {}
func (*EmptyParameter) Descriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{0}
}
func (m *EmptyParameter) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EmptyParameter.Unmarshal(m, b)
}
func (m *EmptyParameter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EmptyParameter.Marshal(b, m, deterministic)
}
func (dst *EmptyParameter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmptyParameter.Merge(dst, src)
}
func (m *EmptyParameter) XXX_Size() int {
	return xxx_messageInfo_EmptyParameter.Size(m)
}
func (m *EmptyParameter) XXX_DiscardUnknown() {
	xxx_messageInfo_EmptyParameter.DiscardUnknown(m)
}

var xxx_messageInfo_EmptyParameter proto.InternalMessageInfo

// Ack/Nack- Response message with comment
type AckNackResponse struct {
	Acknack              bool     `protobuf:"varint,1,opt,name=acknack,proto3" json:"acknack,omitempty"`
	Comments             string   `protobuf:"bytes,2,opt,name=comments,proto3" json:"comments,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AckNackResponse) Reset()         { *m = AckNackResponse{} }
func (m *AckNackResponse) String() string { return proto.CompactTextString(m) }
func (*AckNackResponse) ProtoMessage()    {}
func (*AckNackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{1}
}
func (m *AckNackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AckNackResponse.Unmarshal(m, b)
}
func (m *AckNackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AckNackResponse.Marshal(b, m, deterministic)
}
func (dst *AckNackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AckNackResponse.Merge(dst, src)
}
func (m *AckNackResponse) XXX_Size() int {
	return xxx_messageInfo_AckNackResponse.Size(m)
}
func (m *AckNackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AckNackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AckNackResponse proto.InternalMessageInfo

func (m *AckNackResponse) GetAcknack() bool {
	if m != nil {
		return m.Acknack
	}
	return false
}

func (m *AckNackResponse) GetComments() string {
	if m != nil {
		return m.Comments
	}
	return ""
}

type HaltPaymentRequest struct {
	Haltpayment          bool     `protobuf:"varint,1,opt,name=haltpayment,proto3" json:"haltpayment,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HaltPaymentRequest) Reset()         { *m = HaltPaymentRequest{} }
func (m *HaltPaymentRequest) String() string { return proto.CompactTextString(m) }
func (*HaltPaymentRequest) ProtoMessage()    {}
func (*HaltPaymentRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{2}
}
func (m *HaltPaymentRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HaltPaymentRequest.Unmarshal(m, b)
}
func (m *HaltPaymentRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HaltPaymentRequest.Marshal(b, m, deterministic)
}
func (dst *HaltPaymentRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HaltPaymentRequest.Merge(dst, src)
}
func (m *HaltPaymentRequest) XXX_Size() int {
	return xxx_messageInfo_HaltPaymentRequest.Size(m)
}
func (m *HaltPaymentRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HaltPaymentRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HaltPaymentRequest proto.InternalMessageInfo

func (m *HaltPaymentRequest) GetHaltpayment() bool {
	if m != nil {
		return m.Haltpayment
	}
	return false
}

type Price_UI struct {
	Acknack                   bool      `protobuf:"varint,1,opt,name=acknack,proto3" json:"acknack,omitempty"`
	Comments                  string    `protobuf:"bytes,2,opt,name=comments,proto3" json:"comments,omitempty"`
	SpeedAmountSatoshi        int64     `protobuf:"varint,3,opt,name=speed_amount_satoshi,json=speedAmountSatoshi,proto3" json:"speed_amount_satoshi,omitempty"`
	AccelerationAmountSatoshi int64     `protobuf:"varint,4,opt,name=acceleration_amount_satoshi,json=accelerationAmountSatoshi,proto3" json:"acceleration_amount_satoshi,omitempty"`
	TimeAmountSatoshi         int64     `protobuf:"varint,5,opt,name=time_amount_satoshi,json=timeAmountSatoshi,proto3" json:"time_amount_satoshi,omitempty"`
	SpeedAmountSek            float32   `protobuf:"fixed32,6,opt,name=speed_amount_sek,json=speedAmountSek,proto3" json:"speed_amount_sek,omitempty"`
	AccelerationAmountSek     float32   `protobuf:"fixed32,7,opt,name=acceleration_amount_sek,json=accelerationAmountSek,proto3" json:"acceleration_amount_sek,omitempty"`
	TimeAmountSek             float32   `protobuf:"fixed32,8,opt,name=time_amount_sek,json=timeAmountSek,proto3" json:"time_amount_sek,omitempty"`
	Timeunit                  TimeUnit  `protobuf:"varint,9,opt,name=timeunit,proto3,enum=customer_ui_api.TimeUnit" json:"timeunit,omitempty"`
	PaymentRequestInterval    int32     `protobuf:"varint,10,opt,name=paymentRequestInterval,proto3" json:"paymentRequestInterval,omitempty"`
	Priceunit                 PriceUnit `protobuf:"varint,11,opt,name=priceunit,proto3,enum=customer_ui_api.PriceUnit" json:"priceunit,omitempty"`
	XXX_NoUnkeyedLiteral      struct{}  `json:"-"`
	XXX_unrecognized          []byte    `json:"-"`
	XXX_sizecache             int32     `json:"-"`
}

func (m *Price_UI) Reset()         { *m = Price_UI{} }
func (m *Price_UI) String() string { return proto.CompactTextString(m) }
func (*Price_UI) ProtoMessage()    {}
func (*Price_UI) Descriptor() ([]byte, []int) {
	return fileDescriptor_customer_ui_api_c0d3824182ce7979, []int{3}
}
func (m *Price_UI) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Price_UI.Unmarshal(m, b)
}
func (m *Price_UI) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Price_UI.Marshal(b, m, deterministic)
}
func (dst *Price_UI) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Price_UI.Merge(dst, src)
}
func (m *Price_UI) XXX_Size() int {
	return xxx_messageInfo_Price_UI.Size(m)
}
func (m *Price_UI) XXX_DiscardUnknown() {
	xxx_messageInfo_Price_UI.DiscardUnknown(m)
}

var xxx_messageInfo_Price_UI proto.InternalMessageInfo

func (m *Price_UI) GetAcknack() bool {
	if m != nil {
		return m.Acknack
	}
	return false
}

func (m *Price_UI) GetComments() string {
	if m != nil {
		return m.Comments
	}
	return ""
}

func (m *Price_UI) GetSpeedAmountSatoshi() int64 {
	if m != nil {
		return m.SpeedAmountSatoshi
	}
	return 0
}

func (m *Price_UI) GetAccelerationAmountSatoshi() int64 {
	if m != nil {
		return m.AccelerationAmountSatoshi
	}
	return 0
}

func (m *Price_UI) GetTimeAmountSatoshi() int64 {
	if m != nil {
		return m.TimeAmountSatoshi
	}
	return 0
}

func (m *Price_UI) GetSpeedAmountSek() float32 {
	if m != nil {
		return m.SpeedAmountSek
	}
	return 0
}

func (m *Price_UI) GetAccelerationAmountSek() float32 {
	if m != nil {
		return m.AccelerationAmountSek
	}
	return 0
}

func (m *Price_UI) GetTimeAmountSek() float32 {
	if m != nil {
		return m.TimeAmountSek
	}
	return 0
}

func (m *Price_UI) GetTimeunit() TimeUnit {
	if m != nil {
		return m.Timeunit
	}
	return TimeUnit_SecondsBetweenPaymentmentRequests
}

func (m *Price_UI) GetPaymentRequestInterval() int32 {
	if m != nil {
		return m.PaymentRequestInterval
	}
	return 0
}

func (m *Price_UI) GetPriceunit() PriceUnit {
	if m != nil {
		return m.Priceunit
	}
	return PriceUnit_SatoshiPerSecond
}

func init() {
	proto.RegisterType((*EmptyParameter)(nil), "customer_ui_api.EmptyParameter")
	proto.RegisterType((*AckNackResponse)(nil), "customer_ui_api.AckNackResponse")
	proto.RegisterType((*HaltPaymentRequest)(nil), "customer_ui_api.HaltPaymentRequest")
	proto.RegisterType((*Price_UI)(nil), "customer_ui_api.Price_UI")
	proto.RegisterEnum("customer_ui_api.PriceUnit", PriceUnit_name, PriceUnit_value)
	proto.RegisterEnum("customer_ui_api.TimeUnit", TimeUnit_name, TimeUnit_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Customer_UIClient is the client API for Customer_UI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type Customer_UIClient interface {
	// Ask taxi for Price
	AskTaxiForPrice(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*Price_UI, error)
	// Accept price from Taxi
	AcceptPrice(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*AckNackResponse, error)
	// Halt payment of incoming paymentRequests
	HaltPayments(ctx context.Context, in *HaltPaymentRequest, opts ...grpc.CallOption) (*AckNackResponse, error)
	// Leave Taxi
	LeaveTaxi(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*AckNackResponse, error)
}

type customer_UIClient struct {
	cc *grpc.ClientConn
}

func NewCustomer_UIClient(cc *grpc.ClientConn) Customer_UIClient {
	return &customer_UIClient{cc}
}

func (c *customer_UIClient) AskTaxiForPrice(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*Price_UI, error) {
	out := new(Price_UI)
	err := c.cc.Invoke(ctx, "/customer_ui_api.Customer_UI/AskTaxiForPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customer_UIClient) AcceptPrice(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*AckNackResponse, error) {
	out := new(AckNackResponse)
	err := c.cc.Invoke(ctx, "/customer_ui_api.Customer_UI/AcceptPrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customer_UIClient) HaltPayments(ctx context.Context, in *HaltPaymentRequest, opts ...grpc.CallOption) (*AckNackResponse, error) {
	out := new(AckNackResponse)
	err := c.cc.Invoke(ctx, "/customer_ui_api.Customer_UI/HaltPayments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customer_UIClient) LeaveTaxi(ctx context.Context, in *EmptyParameter, opts ...grpc.CallOption) (*AckNackResponse, error) {
	out := new(AckNackResponse)
	err := c.cc.Invoke(ctx, "/customer_ui_api.Customer_UI/LeaveTaxi", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Customer_UIServer is the server API for Customer_UI service.
type Customer_UIServer interface {
	// Ask taxi for Price
	AskTaxiForPrice(context.Context, *EmptyParameter) (*Price_UI, error)
	// Accept price from Taxi
	AcceptPrice(context.Context, *EmptyParameter) (*AckNackResponse, error)
	// Halt payment of incoming paymentRequests
	HaltPayments(context.Context, *HaltPaymentRequest) (*AckNackResponse, error)
	// Leave Taxi
	LeaveTaxi(context.Context, *EmptyParameter) (*AckNackResponse, error)
}

func RegisterCustomer_UIServer(s *grpc.Server, srv Customer_UIServer) {
	s.RegisterService(&_Customer_UI_serviceDesc, srv)
}

func _Customer_UI_AskTaxiForPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyParameter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Customer_UIServer).AskTaxiForPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/customer_ui_api.Customer_UI/AskTaxiForPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Customer_UIServer).AskTaxiForPrice(ctx, req.(*EmptyParameter))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_UI_AcceptPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyParameter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Customer_UIServer).AcceptPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/customer_ui_api.Customer_UI/AcceptPrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Customer_UIServer).AcceptPrice(ctx, req.(*EmptyParameter))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_UI_HaltPayments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HaltPaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Customer_UIServer).HaltPayments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/customer_ui_api.Customer_UI/HaltPayments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Customer_UIServer).HaltPayments(ctx, req.(*HaltPaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_UI_LeaveTaxi_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyParameter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(Customer_UIServer).LeaveTaxi(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/customer_ui_api.Customer_UI/LeaveTaxi",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(Customer_UIServer).LeaveTaxi(ctx, req.(*EmptyParameter))
	}
	return interceptor(ctx, in, info, handler)
}

var _Customer_UI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "customer_ui_api.Customer_UI",
	HandlerType: (*Customer_UIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AskTaxiForPrice",
			Handler:    _Customer_UI_AskTaxiForPrice_Handler,
		},
		{
			MethodName: "AcceptPrice",
			Handler:    _Customer_UI_AcceptPrice_Handler,
		},
		{
			MethodName: "HaltPayments",
			Handler:    _Customer_UI_HaltPayments_Handler,
		},
		{
			MethodName: "LeaveTaxi",
			Handler:    _Customer_UI_LeaveTaxi_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "customer_ui_api.proto",
}

func init() {
	proto.RegisterFile("customer_ui_api.proto", fileDescriptor_customer_ui_api_c0d3824182ce7979)
}

var fileDescriptor_customer_ui_api_c0d3824182ce7979 = []byte{
	// 513 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0x5b, 0x6f, 0xd3, 0x4c,
	0x10, 0x8d, 0xd3, 0x9b, 0x33, 0xf9, 0xbe, 0x24, 0x0c, 0x2d, 0x38, 0xe1, 0x01, 0xd7, 0x88, 0xca,
	0xca, 0x43, 0x84, 0x8a, 0x88, 0x78, 0x42, 0x0a, 0x88, 0x4b, 0x24, 0x2e, 0x91, 0xdb, 0x88, 0x27,
	0x64, 0x2d, 0xdb, 0x91, 0xba, 0x5a, 0xdf, 0xf0, 0x6e, 0x0a, 0xfd, 0x21, 0xfc, 0x3c, 0xfe, 0x0b,
	0xca, 0xe6, 0x82, 0x1d, 0x07, 0x51, 0x21, 0x1e, 0xe7, 0xcc, 0x39, 0x33, 0x67, 0x3c, 0x3b, 0x86,
	0x23, 0x3e, 0x53, 0x3a, 0x8d, 0x29, 0x0f, 0x67, 0x22, 0x64, 0x99, 0x18, 0x64, 0x79, 0xaa, 0x53,
	0x6c, 0x6f, 0xc0, 0x5e, 0x07, 0x5a, 0x2f, 0xe3, 0x4c, 0x5f, 0x4f, 0x58, 0xce, 0x62, 0xd2, 0x94,
	0x7b, 0xaf, 0xa1, 0x3d, 0xe2, 0xf2, 0x3d, 0xe3, 0x32, 0x20, 0x95, 0xa5, 0x89, 0x22, 0x74, 0xe0,
	0x80, 0x71, 0x99, 0x30, 0x2e, 0x1d, 0xcb, 0xb5, 0x7c, 0x3b, 0x58, 0x85, 0xd8, 0x03, 0x9b, 0xa7,
	0x71, 0x4c, 0x89, 0x56, 0x4e, 0xdd, 0xb5, 0xfc, 0x46, 0xb0, 0x8e, 0xbd, 0x21, 0xe0, 0x1b, 0x16,
	0xe9, 0x09, 0xbb, 0x9e, 0xc7, 0x01, 0x7d, 0x99, 0x91, 0xd2, 0xe8, 0x42, 0xf3, 0x92, 0x45, 0x3a,
	0x5b, 0xa0, 0xcb, 0x7a, 0x45, 0xc8, 0xfb, 0xbe, 0x0b, 0xf6, 0x24, 0x17, 0x9c, 0xc2, 0xe9, 0xf8,
	0xef, 0x5a, 0xe3, 0x23, 0x38, 0x54, 0x19, 0xd1, 0x45, 0xc8, 0xe2, 0x74, 0x96, 0xe8, 0x50, 0x31,
	0x9d, 0xaa, 0x4b, 0xe1, 0xec, 0xb8, 0x96, 0xbf, 0x13, 0xa0, 0xc9, 0x8d, 0x4c, 0xea, 0x6c, 0x91,
	0xc1, 0x67, 0x70, 0x8f, 0x71, 0x4e, 0x11, 0xe5, 0x4c, 0x8b, 0x34, 0xd9, 0x14, 0xee, 0x1a, 0x61,
	0xb7, 0x48, 0x29, 0xeb, 0x07, 0x70, 0x5b, 0x8b, 0x98, 0x36, 0x75, 0x7b, 0x46, 0x77, 0x6b, 0x9e,
	0x2a, 0xf3, 0x7d, 0xe8, 0x94, 0x1d, 0x92, 0x74, 0xf6, 0x5d, 0xcb, 0xaf, 0x07, 0xad, 0xa2, 0x3b,
	0x92, 0x38, 0x84, 0xbb, 0x5b, 0x9d, 0x91, 0x74, 0x0e, 0x8c, 0xe0, 0x68, 0x8b, 0x2b, 0x92, 0x78,
	0x02, 0xed, 0x92, 0x23, 0x92, 0x8e, 0x6d, 0xf8, 0xff, 0x17, 0xdc, 0x90, 0xc4, 0x27, 0x60, 0xcf,
	0x81, 0x59, 0x22, 0xb4, 0xd3, 0x70, 0x2d, 0xbf, 0x75, 0xda, 0x1d, 0x6c, 0x3e, 0x9e, 0x73, 0x11,
	0xd3, 0x34, 0x11, 0x3a, 0x58, 0x53, 0x71, 0x08, 0x77, 0xb2, 0xd2, 0x66, 0xc7, 0x89, 0xa6, 0xfc,
	0x8a, 0x45, 0x0e, 0xb8, 0x96, 0xbf, 0x17, 0xfc, 0x26, 0x8b, 0x4f, 0xa1, 0x91, 0xcd, 0x97, 0x6b,
	0xfa, 0x35, 0x4d, 0xbf, 0x5e, 0xa5, 0x9f, 0x59, 0xbf, 0x69, 0xf8, 0x8b, 0xdc, 0x3f, 0x86, 0xc6,
	0x1a, 0xc7, 0x43, 0xe8, 0x2c, 0x3f, 0xe5, 0x84, 0xf2, 0x33, 0xe2, 0x69, 0x72, 0xd1, 0xa9, 0xf5,
	0x3f, 0x81, 0xbd, 0xb2, 0x8a, 0x0f, 0xe1, 0x78, 0x81, 0xab, 0xe7, 0xa4, 0xbf, 0x12, 0x25, 0xcb,
	0x87, 0x58, 0x30, 0xa5, 0x3a, 0x35, 0xec, 0xc3, 0xc9, 0x3b, 0x11, 0x45, 0xe2, 0xcf, 0x5c, 0xeb,
	0xf4, 0x47, 0x1d, 0x9a, 0x2f, 0x56, 0x56, 0xa7, 0x63, 0xfc, 0x00, 0xed, 0x91, 0x92, 0xe7, 0xec,
	0x9b, 0x78, 0x95, 0xe6, 0xc6, 0x1b, 0xde, 0xaf, 0xcc, 0x52, 0x3e, 0xaf, 0x5e, 0x77, 0xfb, 0xb0,
	0xe1, 0x74, 0xec, 0xd5, 0x30, 0x80, 0xe6, 0x88, 0x73, 0xca, 0xf4, 0x0d, 0x8b, 0xb9, 0x15, 0xc2,
	0xc6, 0xe9, 0x7a, 0x35, 0xfc, 0x08, 0xff, 0x15, 0xce, 0x50, 0xe1, 0x83, 0x8a, 0xa6, 0x7a, 0xa5,
	0x37, 0x2a, 0x3c, 0x81, 0xc6, 0x5b, 0x62, 0x57, 0x34, 0x9f, 0xff, 0x9f, 0x58, 0xfd, 0xbc, 0x6f,
	0x7e, 0x52, 0x8f, 0x7f, 0x06, 0x00, 0x00, 0xff, 0xff, 0x7e, 0x35, 0x14, 0x79, 0xbd, 0x04, 0x00,
	0x00,
}
