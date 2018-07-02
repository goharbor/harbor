// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hapi/release/test_run.proto

package release

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TestRun_Status int32

const (
	TestRun_UNKNOWN TestRun_Status = 0
	TestRun_SUCCESS TestRun_Status = 1
	TestRun_FAILURE TestRun_Status = 2
	TestRun_RUNNING TestRun_Status = 3
)

var TestRun_Status_name = map[int32]string{
	0: "UNKNOWN",
	1: "SUCCESS",
	2: "FAILURE",
	3: "RUNNING",
}
var TestRun_Status_value = map[string]int32{
	"UNKNOWN": 0,
	"SUCCESS": 1,
	"FAILURE": 2,
	"RUNNING": 3,
}

func (x TestRun_Status) String() string {
	return proto.EnumName(TestRun_Status_name, int32(x))
}
func (TestRun_Status) EnumDescriptor() ([]byte, []int) { return fileDescriptor4, []int{0, 0} }

type TestRun struct {
	Name        string                     `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Status      TestRun_Status             `protobuf:"varint,2,opt,name=status,enum=hapi.release.TestRun_Status" json:"status,omitempty"`
	Info        string                     `protobuf:"bytes,3,opt,name=info" json:"info,omitempty"`
	StartedAt   *google_protobuf.Timestamp `protobuf:"bytes,4,opt,name=started_at,json=startedAt" json:"started_at,omitempty"`
	CompletedAt *google_protobuf.Timestamp `protobuf:"bytes,5,opt,name=completed_at,json=completedAt" json:"completed_at,omitempty"`
}

func (m *TestRun) Reset()                    { *m = TestRun{} }
func (m *TestRun) String() string            { return proto.CompactTextString(m) }
func (*TestRun) ProtoMessage()               {}
func (*TestRun) Descriptor() ([]byte, []int) { return fileDescriptor4, []int{0} }

func (m *TestRun) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TestRun) GetStatus() TestRun_Status {
	if m != nil {
		return m.Status
	}
	return TestRun_UNKNOWN
}

func (m *TestRun) GetInfo() string {
	if m != nil {
		return m.Info
	}
	return ""
}

func (m *TestRun) GetStartedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.StartedAt
	}
	return nil
}

func (m *TestRun) GetCompletedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.CompletedAt
	}
	return nil
}

func init() {
	proto.RegisterType((*TestRun)(nil), "hapi.release.TestRun")
	proto.RegisterEnum("hapi.release.TestRun_Status", TestRun_Status_name, TestRun_Status_value)
}

func init() { proto.RegisterFile("hapi/release/test_run.proto", fileDescriptor4) }

var fileDescriptor4 = []byte{
	// 274 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x8f, 0xc1, 0x4b, 0xfb, 0x30,
	0x1c, 0xc5, 0x7f, 0xe9, 0xf6, 0x6b, 0x69, 0x3a, 0xa4, 0xe4, 0x54, 0xa6, 0x60, 0xd9, 0xa9, 0xa7,
	0x14, 0xa6, 0x17, 0x41, 0x0f, 0x75, 0x4c, 0x19, 0x4a, 0x84, 0x74, 0x45, 0xf0, 0x32, 0x32, 0xcd,
	0x66, 0xa1, 0x6d, 0x4a, 0xf3, 0xed, 0xdf, 0xe3, 0xbf, 0x2a, 0x69, 0x33, 0xf1, 0xe6, 0xed, 0xfb,
	0x78, 0x9f, 0xf7, 0xf2, 0x82, 0xcf, 0x3f, 0x45, 0x5b, 0xa6, 0x9d, 0xac, 0xa4, 0xd0, 0x32, 0x05,
	0xa9, 0x61, 0xd7, 0xf5, 0x0d, 0x6d, 0x3b, 0x05, 0x8a, 0xcc, 0x8c, 0x49, 0xad, 0x39, 0xbf, 0x3c,
	0x2a, 0x75, 0xac, 0x64, 0x3a, 0x78, 0xfb, 0xfe, 0x90, 0x42, 0x59, 0x4b, 0x0d, 0xa2, 0x6e, 0x47,
	0x7c, 0xf1, 0xe5, 0x60, 0x6f, 0x2b, 0x35, 0xf0, 0xbe, 0x21, 0x04, 0x4f, 0x1b, 0x51, 0xcb, 0x08,
	0xc5, 0x28, 0xf1, 0xf9, 0x70, 0x93, 0x6b, 0xec, 0x6a, 0x10, 0xd0, 0xeb, 0xc8, 0x89, 0x51, 0x72,
	0xb6, 0xbc, 0xa0, 0xbf, 0xfb, 0xa9, 0x8d, 0xd2, 0x7c, 0x60, 0xb8, 0x65, 0x4d, 0x53, 0xd9, 0x1c,
	0x54, 0x34, 0x19, 0x9b, 0xcc, 0x4d, 0x6e, 0x30, 0xd6, 0x20, 0x3a, 0x90, 0x1f, 0x3b, 0x01, 0xd1,
	0x34, 0x46, 0x49, 0xb0, 0x9c, 0xd3, 0x71, 0x1f, 0x3d, 0xed, 0xa3, 0xdb, 0xd3, 0x3e, 0xee, 0x5b,
	0x3a, 0x03, 0x72, 0x87, 0x67, 0xef, 0xaa, 0x6e, 0x2b, 0x69, 0xc3, 0xff, 0xff, 0x0c, 0x07, 0x3f,
	0x7c, 0x06, 0x8b, 0x5b, 0xec, 0x8e, 0xfb, 0x48, 0x80, 0xbd, 0x82, 0x3d, 0xb1, 0x97, 0x57, 0x16,
	0xfe, 0x33, 0x22, 0x2f, 0x56, 0xab, 0x75, 0x9e, 0x87, 0xc8, 0x88, 0x87, 0x6c, 0xf3, 0x5c, 0xf0,
	0x75, 0xe8, 0x18, 0xc1, 0x0b, 0xc6, 0x36, 0xec, 0x31, 0x9c, 0xdc, 0xfb, 0x6f, 0x9e, 0xfd, 0xed,
	0xde, 0x1d, 0x5e, 0xba, 0xfa, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x31, 0x86, 0x46, 0xdb, 0x81, 0x01,
	0x00, 0x00,
}
