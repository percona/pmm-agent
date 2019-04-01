// Code generated by protoc-gen-go. DO NOT EDIT.
// source: qanpb/profile.proto

package qanpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import timestamp "github.com/golang/protobuf/ptypes/timestamp"
import _ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
import _ "google.golang.org/genproto/googleapis/api/annotations"

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

// ReportRequest defines filtering of metrics report for db server or other dimentions.
type ReportRequest struct {
	PeriodStartFrom      *timestamp.Timestamp   `protobuf:"bytes,1,opt,name=period_start_from,json=periodStartFrom,proto3" json:"period_start_from,omitempty"`
	PeriodStartTo        *timestamp.Timestamp   `protobuf:"bytes,2,opt,name=period_start_to,json=periodStartTo,proto3" json:"period_start_to,omitempty"`
	GroupBy              string                 `protobuf:"bytes,3,opt,name=group_by,json=groupBy,proto3" json:"group_by,omitempty"`
	Labels               []*ReportMapFieldEntry `protobuf:"bytes,4,rep,name=labels,proto3" json:"labels,omitempty"`
	Columns              []string               `protobuf:"bytes,5,rep,name=columns,proto3" json:"columns,omitempty"`
	OrderBy              string                 `protobuf:"bytes,6,opt,name=order_by,json=orderBy,proto3" json:"order_by,omitempty"`
	Offset               uint32                 `protobuf:"varint,7,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit                uint32                 `protobuf:"varint,8,opt,name=limit,proto3" json:"limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ReportRequest) Reset()         { *m = ReportRequest{} }
func (m *ReportRequest) String() string { return proto.CompactTextString(m) }
func (*ReportRequest) ProtoMessage()    {}
func (*ReportRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{0}
}
func (m *ReportRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportRequest.Unmarshal(m, b)
}
func (m *ReportRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportRequest.Marshal(b, m, deterministic)
}
func (dst *ReportRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportRequest.Merge(dst, src)
}
func (m *ReportRequest) XXX_Size() int {
	return xxx_messageInfo_ReportRequest.Size(m)
}
func (m *ReportRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ReportRequest proto.InternalMessageInfo

func (m *ReportRequest) GetPeriodStartFrom() *timestamp.Timestamp {
	if m != nil {
		return m.PeriodStartFrom
	}
	return nil
}

func (m *ReportRequest) GetPeriodStartTo() *timestamp.Timestamp {
	if m != nil {
		return m.PeriodStartTo
	}
	return nil
}

func (m *ReportRequest) GetGroupBy() string {
	if m != nil {
		return m.GroupBy
	}
	return ""
}

func (m *ReportRequest) GetLabels() []*ReportMapFieldEntry {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *ReportRequest) GetColumns() []string {
	if m != nil {
		return m.Columns
	}
	return nil
}

func (m *ReportRequest) GetOrderBy() string {
	if m != nil {
		return m.OrderBy
	}
	return ""
}

func (m *ReportRequest) GetOffset() uint32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *ReportRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

// ReportMapFieldEntry allows to pass labels/dimentions in form like {"d_server": ["db1", "db2"...]}.
type ReportMapFieldEntry struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                []string `protobuf:"bytes,2,rep,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportMapFieldEntry) Reset()         { *m = ReportMapFieldEntry{} }
func (m *ReportMapFieldEntry) String() string { return proto.CompactTextString(m) }
func (*ReportMapFieldEntry) ProtoMessage()    {}
func (*ReportMapFieldEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{1}
}
func (m *ReportMapFieldEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportMapFieldEntry.Unmarshal(m, b)
}
func (m *ReportMapFieldEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportMapFieldEntry.Marshal(b, m, deterministic)
}
func (dst *ReportMapFieldEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportMapFieldEntry.Merge(dst, src)
}
func (m *ReportMapFieldEntry) XXX_Size() int {
	return xxx_messageInfo_ReportMapFieldEntry.Size(m)
}
func (m *ReportMapFieldEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportMapFieldEntry.DiscardUnknown(m)
}

var xxx_messageInfo_ReportMapFieldEntry proto.InternalMessageInfo

func (m *ReportMapFieldEntry) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ReportMapFieldEntry) GetValue() []string {
	if m != nil {
		return m.Value
	}
	return nil
}

// ReportReply is list of reports per quieryids, hosts etc.
type ReportReply struct {
	TotalRows            uint32   `protobuf:"varint,1,opt,name=total_rows,json=totalRows,proto3" json:"total_rows,omitempty"`
	Offset               uint32   `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit                uint32   `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	Rows                 []*Row   `protobuf:"bytes,4,rep,name=rows,proto3" json:"rows,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ReportReply) Reset()         { *m = ReportReply{} }
func (m *ReportReply) String() string { return proto.CompactTextString(m) }
func (*ReportReply) ProtoMessage()    {}
func (*ReportReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{2}
}
func (m *ReportReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReportReply.Unmarshal(m, b)
}
func (m *ReportReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReportReply.Marshal(b, m, deterministic)
}
func (dst *ReportReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReportReply.Merge(dst, src)
}
func (m *ReportReply) XXX_Size() int {
	return xxx_messageInfo_ReportReply.Size(m)
}
func (m *ReportReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ReportReply.DiscardUnknown(m)
}

var xxx_messageInfo_ReportReply proto.InternalMessageInfo

func (m *ReportReply) GetTotalRows() uint32 {
	if m != nil {
		return m.TotalRows
	}
	return 0
}

func (m *ReportReply) GetOffset() uint32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *ReportReply) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *ReportReply) GetRows() []*Row {
	if m != nil {
		return m.Rows
	}
	return nil
}

// Row define metrics for selected dimention.
type Row struct {
	Rank                 uint32             `protobuf:"varint,1,opt,name=rank,proto3" json:"rank,omitempty"`
	Dimension            string             `protobuf:"bytes,2,opt,name=dimension,proto3" json:"dimension,omitempty"`
	Metrics              map[string]*Metric `protobuf:"bytes,3,rep,name=metrics,proto3" json:"metrics,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Sparkline            []*Point           `protobuf:"bytes,4,rep,name=sparkline,proto3" json:"sparkline,omitempty"`
	Fingerprint          string             `protobuf:"bytes,5,opt,name=fingerprint,proto3" json:"fingerprint,omitempty"`
	NumQueries           uint32             `protobuf:"varint,6,opt,name=num_queries,json=numQueries,proto3" json:"num_queries,omitempty"`
	Qps                  float32            `protobuf:"fixed32,7,opt,name=qps,proto3" json:"qps,omitempty"`
	Load                 float32            `protobuf:"fixed32,8,opt,name=load,proto3" json:"load,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Row) Reset()         { *m = Row{} }
func (m *Row) String() string { return proto.CompactTextString(m) }
func (*Row) ProtoMessage()    {}
func (*Row) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{3}
}
func (m *Row) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Row.Unmarshal(m, b)
}
func (m *Row) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Row.Marshal(b, m, deterministic)
}
func (dst *Row) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Row.Merge(dst, src)
}
func (m *Row) XXX_Size() int {
	return xxx_messageInfo_Row.Size(m)
}
func (m *Row) XXX_DiscardUnknown() {
	xxx_messageInfo_Row.DiscardUnknown(m)
}

var xxx_messageInfo_Row proto.InternalMessageInfo

func (m *Row) GetRank() uint32 {
	if m != nil {
		return m.Rank
	}
	return 0
}

func (m *Row) GetDimension() string {
	if m != nil {
		return m.Dimension
	}
	return ""
}

func (m *Row) GetMetrics() map[string]*Metric {
	if m != nil {
		return m.Metrics
	}
	return nil
}

func (m *Row) GetSparkline() []*Point {
	if m != nil {
		return m.Sparkline
	}
	return nil
}

func (m *Row) GetFingerprint() string {
	if m != nil {
		return m.Fingerprint
	}
	return ""
}

func (m *Row) GetNumQueries() uint32 {
	if m != nil {
		return m.NumQueries
	}
	return 0
}

func (m *Row) GetQps() float32 {
	if m != nil {
		return m.Qps
	}
	return 0
}

func (m *Row) GetLoad() float32 {
	if m != nil {
		return m.Load
	}
	return 0
}

// Metric cell.
type Metric struct {
	Stats                *Stat    `protobuf:"bytes,1,opt,name=stats,proto3" json:"stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Metric) Reset()         { *m = Metric{} }
func (m *Metric) String() string { return proto.CompactTextString(m) }
func (*Metric) ProtoMessage()    {}
func (*Metric) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{4}
}
func (m *Metric) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metric.Unmarshal(m, b)
}
func (m *Metric) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metric.Marshal(b, m, deterministic)
}
func (dst *Metric) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metric.Merge(dst, src)
}
func (m *Metric) XXX_Size() int {
	return xxx_messageInfo_Metric.Size(m)
}
func (m *Metric) XXX_DiscardUnknown() {
	xxx_messageInfo_Metric.DiscardUnknown(m)
}

var xxx_messageInfo_Metric proto.InternalMessageInfo

func (m *Metric) GetStats() *Stat {
	if m != nil {
		return m.Stats
	}
	return nil
}

// Stat is statistics of specific metric.
type Stat struct {
	Rate                 float32  `protobuf:"fixed32,1,opt,name=rate,proto3" json:"rate,omitempty"`
	Cnt                  float32  `protobuf:"fixed32,2,opt,name=cnt,proto3" json:"cnt,omitempty"`
	Sum                  float32  `protobuf:"fixed32,3,opt,name=sum,proto3" json:"sum,omitempty"`
	Min                  float32  `protobuf:"fixed32,4,opt,name=min,proto3" json:"min,omitempty"`
	Max                  float32  `protobuf:"fixed32,5,opt,name=max,proto3" json:"max,omitempty"`
	P99                  float32  `protobuf:"fixed32,6,opt,name=p99,proto3" json:"p99,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Stat) Reset()         { *m = Stat{} }
func (m *Stat) String() string { return proto.CompactTextString(m) }
func (*Stat) ProtoMessage()    {}
func (*Stat) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{5}
}
func (m *Stat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stat.Unmarshal(m, b)
}
func (m *Stat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stat.Marshal(b, m, deterministic)
}
func (dst *Stat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stat.Merge(dst, src)
}
func (m *Stat) XXX_Size() int {
	return xxx_messageInfo_Stat.Size(m)
}
func (m *Stat) XXX_DiscardUnknown() {
	xxx_messageInfo_Stat.DiscardUnknown(m)
}

var xxx_messageInfo_Stat proto.InternalMessageInfo

func (m *Stat) GetRate() float32 {
	if m != nil {
		return m.Rate
	}
	return 0
}

func (m *Stat) GetCnt() float32 {
	if m != nil {
		return m.Cnt
	}
	return 0
}

func (m *Stat) GetSum() float32 {
	if m != nil {
		return m.Sum
	}
	return 0
}

func (m *Stat) GetMin() float32 {
	if m != nil {
		return m.Min
	}
	return 0
}

func (m *Stat) GetMax() float32 {
	if m != nil {
		return m.Max
	}
	return 0
}

func (m *Stat) GetP99() float32 {
	if m != nil {
		return m.P99
	}
	return 0
}

// Point contains values that represents abscissa (time) and ordinate (volume etc.)
// of every point in a coordinate system of Sparklines.
type Point struct {
	Values               map[string]float32 `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"fixed32,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Point) Reset()         { *m = Point{} }
func (m *Point) String() string { return proto.CompactTextString(m) }
func (*Point) ProtoMessage()    {}
func (*Point) Descriptor() ([]byte, []int) {
	return fileDescriptor_profile_2426f5e2ce0d2233, []int{6}
}
func (m *Point) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Point.Unmarshal(m, b)
}
func (m *Point) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Point.Marshal(b, m, deterministic)
}
func (dst *Point) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Point.Merge(dst, src)
}
func (m *Point) XXX_Size() int {
	return xxx_messageInfo_Point.Size(m)
}
func (m *Point) XXX_DiscardUnknown() {
	xxx_messageInfo_Point.DiscardUnknown(m)
}

var xxx_messageInfo_Point proto.InternalMessageInfo

func (m *Point) GetValues() map[string]float32 {
	if m != nil {
		return m.Values
	}
	return nil
}

func init() {
	proto.RegisterType((*ReportRequest)(nil), "qan.ReportRequest")
	proto.RegisterType((*ReportMapFieldEntry)(nil), "qan.ReportMapFieldEntry")
	proto.RegisterType((*ReportReply)(nil), "qan.ReportReply")
	proto.RegisterType((*Row)(nil), "qan.Row")
	proto.RegisterMapType((map[string]*Metric)(nil), "qan.Row.MetricsEntry")
	proto.RegisterType((*Metric)(nil), "qan.Metric")
	proto.RegisterType((*Stat)(nil), "qan.Stat")
	proto.RegisterType((*Point)(nil), "qan.Point")
	proto.RegisterMapType((map[string]float32)(nil), "qan.Point.ValuesEntry")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ProfileClient is the client API for Profile service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ProfileClient interface {
	// GetReport returns list of metrics group by queryid or other dimentions.
	GetReport(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportReply, error)
}

type profileClient struct {
	cc *grpc.ClientConn
}

func NewProfileClient(cc *grpc.ClientConn) ProfileClient {
	return &profileClient{cc}
}

func (c *profileClient) GetReport(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportReply, error) {
	out := new(ReportReply)
	err := c.cc.Invoke(ctx, "/qan.Profile/GetReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProfileServer is the server API for Profile service.
type ProfileServer interface {
	// GetReport returns list of metrics group by queryid or other dimentions.
	GetReport(context.Context, *ReportRequest) (*ReportReply, error)
}

func RegisterProfileServer(s *grpc.Server, srv ProfileServer) {
	s.RegisterService(&_Profile_serviceDesc, srv)
}

func _Profile_GetReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProfileServer).GetReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/qan.Profile/GetReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProfileServer).GetReport(ctx, req.(*ReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Profile_serviceDesc = grpc.ServiceDesc{
	ServiceName: "qan.Profile",
	HandlerType: (*ProfileServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetReport",
			Handler:    _Profile_GetReport_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "qanpb/profile.proto",
}

func init() { proto.RegisterFile("qanpb/profile.proto", fileDescriptor_profile_2426f5e2ce0d2233) }

var fileDescriptor_profile_2426f5e2ce0d2233 = []byte{
	// 776 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x54, 0xcd, 0x6e, 0xe3, 0x36,
	0x10, 0x86, 0x25, 0xff, 0x44, 0xe3, 0x1a, 0xbb, 0xcb, 0x6d, 0x03, 0xd5, 0x48, 0x11, 0x55, 0xed,
	0xc1, 0x2d, 0xba, 0x52, 0x9b, 0x5e, 0x9a, 0x00, 0x3d, 0x6c, 0x80, 0x24, 0xa7, 0x45, 0x77, 0xb9,
	0x8b, 0x1e, 0xf6, 0x62, 0xd0, 0x31, 0x65, 0x10, 0x91, 0x48, 0x9a, 0xa4, 0xd6, 0xf1, 0xb5, 0x8f,
	0xd0, 0xbe, 0x52, 0x6f, 0x3d, 0xf6, 0x15, 0xfa, 0x20, 0x05, 0x87, 0x72, 0xed, 0xa0, 0x41, 0x7b,
	0xd2, 0xcc, 0xc7, 0x99, 0xf9, 0x46, 0xf3, 0x91, 0x03, 0xcf, 0xd7, 0x4c, 0xea, 0x45, 0xa9, 0x8d,
	0xaa, 0x44, 0xcd, 0x0b, 0x6d, 0x94, 0x53, 0x24, 0x5e, 0x33, 0x39, 0x3d, 0x59, 0x29, 0xb5, 0xaa,
	0x79, 0xc9, 0xb4, 0x28, 0x99, 0x94, 0xca, 0x31, 0x27, 0x94, 0xb4, 0x21, 0x64, 0x7a, 0xda, 0x9d,
	0xa2, 0xb7, 0x68, 0xab, 0xd2, 0x89, 0x86, 0x5b, 0xc7, 0x1a, 0xdd, 0x05, 0x7c, 0x83, 0x9f, 0xdb,
	0x17, 0x2b, 0x2e, 0x5f, 0xd8, 0x0d, 0x5b, 0xad, 0xb8, 0x29, 0x95, 0xc6, 0x12, 0xff, 0x2e, 0x97,
	0xff, 0x11, 0xc1, 0x84, 0x72, 0xad, 0x8c, 0xa3, 0x7c, 0xdd, 0x72, 0xeb, 0xc8, 0x35, 0x3c, 0xd3,
	0xdc, 0x08, 0xb5, 0x9c, 0x5b, 0xc7, 0x8c, 0x9b, 0x57, 0x46, 0x35, 0x69, 0x2f, 0xeb, 0xcd, 0xc6,
	0x67, 0xd3, 0x22, 0x90, 0x17, 0x3b, 0xf2, 0xe2, 0xdd, 0x8e, 0x9c, 0x3e, 0x09, 0x49, 0x6f, 0x7d,
	0xce, 0xb5, 0x51, 0x0d, 0xb9, 0x84, 0x27, 0x0f, 0xea, 0x38, 0x95, 0x46, 0xff, 0x5b, 0x65, 0x72,
	0x50, 0xe5, 0x9d, 0x22, 0x9f, 0xc2, 0xd1, 0xca, 0xa8, 0x56, 0xcf, 0x17, 0xdb, 0x34, 0xce, 0x7a,
	0xb3, 0x84, 0x8e, 0xd0, 0xbf, 0xdc, 0x92, 0x6f, 0x61, 0x58, 0xb3, 0x05, 0xaf, 0x6d, 0xda, 0xcf,
	0xe2, 0xd9, 0xf8, 0x2c, 0x2d, 0xd6, 0x4c, 0x16, 0xe1, 0x57, 0x5e, 0x31, 0x7d, 0x2d, 0x78, 0xbd,
	0xbc, 0x92, 0xce, 0x6c, 0x69, 0x17, 0x47, 0x52, 0x18, 0xdd, 0xaa, 0xba, 0x6d, 0xa4, 0x4d, 0x07,
	0x59, 0xec, 0x6b, 0x75, 0xae, 0xa7, 0x51, 0x66, 0xc9, 0x8d, 0xa7, 0x19, 0x06, 0x1a, 0xf4, 0x2f,
	0xb7, 0xe4, 0x18, 0x86, 0xaa, 0xaa, 0x2c, 0x77, 0xe9, 0x28, 0xeb, 0xcd, 0x26, 0xb4, 0xf3, 0xc8,
	0xc7, 0x30, 0xa8, 0x45, 0x23, 0x5c, 0x7a, 0x84, 0x70, 0x70, 0xf2, 0x1f, 0xe1, 0xf9, 0x23, 0x1d,
	0x90, 0xa7, 0x10, 0xdf, 0xf1, 0x2d, 0x0e, 0x31, 0xa1, 0xde, 0xf4, 0xe9, 0x1f, 0x58, 0xdd, 0xf2,
	0x34, 0xc2, 0x4e, 0x82, 0x93, 0xdf, 0xc3, 0x78, 0xa7, 0x85, 0xae, 0xb7, 0xe4, 0x33, 0x00, 0xa7,
	0x1c, 0xab, 0xe7, 0x46, 0x6d, 0x2c, 0x66, 0x4f, 0x68, 0x82, 0x08, 0x55, 0x1b, 0x7b, 0xd0, 0x5a,
	0xf4, 0x78, 0x6b, 0xf1, 0x41, 0x6b, 0xe4, 0x04, 0xfa, 0x58, 0x26, 0x4c, 0xeb, 0x28, 0x4c, 0x4b,
	0x6d, 0x28, 0xa2, 0xf9, 0xef, 0x11, 0xc4, 0x54, 0x6d, 0x08, 0x81, 0xbe, 0x61, 0xf2, 0xae, 0x23,
	0x43, 0x9b, 0x9c, 0x40, 0xb2, 0x14, 0x0d, 0x97, 0x56, 0x28, 0x89, 0x54, 0x09, 0xdd, 0x03, 0xa4,
	0x84, 0x51, 0xc3, 0x9d, 0x11, 0xb7, 0x36, 0x8d, 0xb1, 0xf4, 0x27, 0xbb, 0xd2, 0xc5, 0xab, 0x80,
	0x07, 0x15, 0x76, 0x51, 0x64, 0x06, 0x89, 0xd5, 0xcc, 0xdc, 0xd5, 0x42, 0xf2, 0xae, 0x1b, 0xc0,
	0x94, 0xd7, 0x4a, 0x48, 0x47, 0xf7, 0x87, 0x24, 0x83, 0x71, 0x25, 0xe4, 0x8a, 0x1b, 0x6d, 0x84,
	0x74, 0xe9, 0x00, 0xa9, 0x0f, 0x21, 0x72, 0x0a, 0x63, 0xd9, 0x36, 0xf3, 0x75, 0xcb, 0x8d, 0xe0,
	0x16, 0xb5, 0x9b, 0x50, 0x90, 0x6d, 0xf3, 0x26, 0x20, 0x7e, 0xf2, 0x6b, 0x6d, 0x51, 0xbb, 0x88,
	0x7a, 0xd3, 0xff, 0x61, 0xad, 0xd8, 0x12, 0x75, 0x8b, 0x28, 0xda, 0xd3, 0x1b, 0xf8, 0xe8, 0xb0,
	0xd7, 0x47, 0xf4, 0xfa, 0x7c, 0xaf, 0x97, 0xbf, 0xc2, 0x63, 0x6c, 0x38, 0xe4, 0x74, 0xe2, 0x5d,
	0x44, 0x3f, 0xf4, 0xf2, 0xaf, 0x60, 0x18, 0x40, 0x72, 0x0a, 0x03, 0xeb, 0x98, 0xb3, 0xdd, 0xcb,
	0x49, 0x30, 0xe1, 0xad, 0x63, 0x8e, 0x06, 0x3c, 0x77, 0xd0, 0xf7, 0x6e, 0x98, 0xb8, 0xe3, 0x18,
	0x17, 0x51, 0xb4, 0x3d, 0xff, 0xad, 0x0c, 0xb2, 0x46, 0xd4, 0x9b, 0x1e, 0xb1, 0x6d, 0x83, 0x8a,
	0x46, 0xd4, 0x9b, 0x1e, 0x69, 0x84, 0x4c, 0xfb, 0x01, 0x69, 0x84, 0x44, 0x84, 0xdd, 0xe3, 0x98,
	0x3c, 0xc2, 0xee, 0x3d, 0xa2, 0xcf, 0xcf, 0x71, 0x2c, 0x11, 0xf5, 0x66, 0x6e, 0x60, 0x80, 0x63,
	0x26, 0x05, 0x0c, 0xb1, 0x6d, 0xdf, 0xa0, 0x97, 0xe0, 0x78, 0x2f, 0x41, 0xf1, 0x33, 0x1e, 0x74,
	0x8f, 0x27, 0x44, 0x4d, 0xcf, 0x61, 0x7c, 0x00, 0xff, 0xf7, 0x8d, 0xf6, 0x6c, 0xfb, 0xa1, 0x9c,
	0xbd, 0x87, 0xd1, 0xeb, 0xb0, 0xe5, 0xc8, 0x4f, 0x90, 0xdc, 0x70, 0x17, 0xee, 0x38, 0x21, 0x07,
	0x2f, 0xb6, 0x5b, 0x3e, 0xd3, 0xa7, 0x0f, 0x30, 0x5d, 0x6f, 0xf3, 0x93, 0x5f, 0xfe, 0xfc, 0xeb,
	0xb7, 0xe8, 0x38, 0x7f, 0x56, 0x7e, 0xf8, 0xae, 0x5c, 0x33, 0x59, 0xfe, 0x53, 0xe0, 0xa2, 0xf7,
	0xf5, 0xe5, 0x9b, 0x5f, 0x5f, 0xde, 0xd0, 0x2b, 0x18, 0x2d, 0x79, 0xc5, 0xda, 0xda, 0x91, 0x0b,
	0x20, 0x2f, 0x65, 0xc6, 0x8d, 0x51, 0x26, 0x33, 0xdc, 0x6a, 0x25, 0x2d, 0x2f, 0xc8, 0x97, 0x90,
	0x4f, 0xb3, 0x2f, 0xca, 0x25, 0xaf, 0x84, 0x14, 0x61, 0x13, 0xe2, 0xf6, 0xbd, 0xf2, 0x71, 0xb4,
	0x0b, 0x7b, 0x3f, 0x40, 0x6c, 0x31, 0xc4, 0xb5, 0xf4, 0xfd, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x34, 0x11, 0xff, 0x0f, 0xa1, 0x05, 0x00, 0x00,
}
