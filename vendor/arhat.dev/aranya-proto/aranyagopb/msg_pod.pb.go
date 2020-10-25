// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: msg_pod.proto

// +build !rt_none

package aranyagopb

import (
	bytes "bytes"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_sortkeys "github.com/gogo/protobuf/sortkeys"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strconv "strconv"
	strings "strings"
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

type PodState int32

const (
	POD_STATE_UNKNOWN   PodState = 0
	POD_STATE_PENDING   PodState = 1
	POD_STATE_RUNNING   PodState = 2
	POD_STATE_SUCCEEDED PodState = 3
	POD_STATE_FAILED    PodState = 4
)

var PodState_name = map[int32]string{
	0: "POD_STATE_UNKNOWN",
	1: "POD_STATE_PENDING",
	2: "POD_STATE_RUNNING",
	3: "POD_STATE_SUCCEEDED",
	4: "POD_STATE_FAILED",
}

var PodState_value = map[string]int32{
	"POD_STATE_UNKNOWN":   0,
	"POD_STATE_PENDING":   1,
	"POD_STATE_RUNNING":   2,
	"POD_STATE_SUCCEEDED": 3,
	"POD_STATE_FAILED":    4,
}

func (PodState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_164584cfac8e9deb, []int{0}
}

type ContainerStatus struct {
	ContainerId string `protobuf:"bytes,1,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	ImageId     string `protobuf:"bytes,2,opt,name=image_id,json=imageId,proto3" json:"image_id,omitempty"`
	// time values in rfc3339nano format
	CreatedAt    string `protobuf:"bytes,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	StartedAt    string `protobuf:"bytes,5,opt,name=started_at,json=startedAt,proto3" json:"started_at,omitempty"`
	FinishedAt   string `protobuf:"bytes,6,opt,name=finished_at,json=finishedAt,proto3" json:"finished_at,omitempty"`
	ExitCode     int32  `protobuf:"varint,7,opt,name=exit_code,json=exitCode,proto3" json:"exit_code,omitempty"`
	RestartCount int32  `protobuf:"varint,8,opt,name=restart_count,json=restartCount,proto3" json:"restart_count,omitempty"`
	Reason       string `protobuf:"bytes,11,opt,name=reason,proto3" json:"reason,omitempty"`
	Message      string `protobuf:"bytes,12,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *ContainerStatus) Reset()      { *m = ContainerStatus{} }
func (*ContainerStatus) ProtoMessage() {}
func (*ContainerStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_164584cfac8e9deb, []int{0}
}
func (m *ContainerStatus) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ContainerStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ContainerStatus.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ContainerStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerStatus.Merge(m, src)
}
func (m *ContainerStatus) XXX_Size() int {
	return m.Size()
}
func (m *ContainerStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerStatus proto.InternalMessageInfo

func (m *ContainerStatus) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

func (m *ContainerStatus) GetImageId() string {
	if m != nil {
		return m.ImageId
	}
	return ""
}

func (m *ContainerStatus) GetCreatedAt() string {
	if m != nil {
		return m.CreatedAt
	}
	return ""
}

func (m *ContainerStatus) GetStartedAt() string {
	if m != nil {
		return m.StartedAt
	}
	return ""
}

func (m *ContainerStatus) GetFinishedAt() string {
	if m != nil {
		return m.FinishedAt
	}
	return ""
}

func (m *ContainerStatus) GetExitCode() int32 {
	if m != nil {
		return m.ExitCode
	}
	return 0
}

func (m *ContainerStatus) GetRestartCount() int32 {
	if m != nil {
		return m.RestartCount
	}
	return 0
}

func (m *ContainerStatus) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *ContainerStatus) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type PodStatusMsg struct {
	// metadata
	Uid string `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	// pod network status, protobuf bytes of abbot proto
	Network []byte `protobuf:"bytes,2,opt,name=network,proto3" json:"network,omitempty"`
	// status
	Containers map[string]*ContainerStatus `protobuf:"bytes,3,rep,name=containers,proto3" json:"containers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *PodStatusMsg) Reset()      { *m = PodStatusMsg{} }
func (*PodStatusMsg) ProtoMessage() {}
func (*PodStatusMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_164584cfac8e9deb, []int{1}
}
func (m *PodStatusMsg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PodStatusMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PodStatusMsg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PodStatusMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PodStatusMsg.Merge(m, src)
}
func (m *PodStatusMsg) XXX_Size() int {
	return m.Size()
}
func (m *PodStatusMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PodStatusMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PodStatusMsg proto.InternalMessageInfo

func (m *PodStatusMsg) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

func (m *PodStatusMsg) GetNetwork() []byte {
	if m != nil {
		return m.Network
	}
	return nil
}

func (m *PodStatusMsg) GetContainers() map[string]*ContainerStatus {
	if m != nil {
		return m.Containers
	}
	return nil
}

type PodStatusListMsg struct {
	Pods []*PodStatusMsg `protobuf:"bytes,1,rep,name=pods,proto3" json:"pods,omitempty"`
}

func (m *PodStatusListMsg) Reset()      { *m = PodStatusListMsg{} }
func (*PodStatusListMsg) ProtoMessage() {}
func (*PodStatusListMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_164584cfac8e9deb, []int{2}
}
func (m *PodStatusListMsg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PodStatusListMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PodStatusListMsg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PodStatusListMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PodStatusListMsg.Merge(m, src)
}
func (m *PodStatusListMsg) XXX_Size() int {
	return m.Size()
}
func (m *PodStatusListMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PodStatusListMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PodStatusListMsg proto.InternalMessageInfo

func (m *PodStatusListMsg) GetPods() []*PodStatusMsg {
	if m != nil {
		return m.Pods
	}
	return nil
}

func init() {
	proto.RegisterEnum("aranya.PodState", PodState_name, PodState_value)
	proto.RegisterType((*ContainerStatus)(nil), "aranya.ContainerStatus")
	proto.RegisterType((*PodStatusMsg)(nil), "aranya.PodStatusMsg")
	proto.RegisterMapType((map[string]*ContainerStatus)(nil), "aranya.PodStatusMsg.ContainersEntry")
	proto.RegisterType((*PodStatusListMsg)(nil), "aranya.PodStatusListMsg")
}

func init() { proto.RegisterFile("msg_pod.proto", fileDescriptor_164584cfac8e9deb) }

var fileDescriptor_164584cfac8e9deb = []byte{
	// 528 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x93, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0xbd, 0xf9, 0x9f, 0x49, 0x2a, 0xcc, 0x52, 0xa8, 0x01, 0xb1, 0xa4, 0x81, 0x43, 0x84,
	0xd4, 0x20, 0x95, 0x0b, 0x42, 0x5c, 0x42, 0x6c, 0x50, 0x44, 0x71, 0x23, 0xa7, 0xa1, 0x12, 0x17,
	0x6b, 0x1b, 0x2f, 0xa9, 0x15, 0xe2, 0x8d, 0xbc, 0x9b, 0x42, 0x2e, 0x88, 0x47, 0xe0, 0x19, 0x38,
	0xf1, 0x28, 0x1c, 0x73, 0xcc, 0x91, 0x38, 0x17, 0x8e, 0x7d, 0x04, 0xe4, 0xb5, 0x93, 0xa6, 0x88,
	0x5b, 0xe6, 0xf7, 0xcd, 0x4e, 0xbe, 0xf9, 0x46, 0x86, 0x9d, 0xb1, 0x18, 0xba, 0x13, 0xee, 0x35,
	0x27, 0x21, 0x97, 0x1c, 0x17, 0x68, 0x48, 0x83, 0x19, 0xad, 0xff, 0xc8, 0xc0, 0x8d, 0x36, 0x0f,
	0x24, 0xf5, 0x03, 0x16, 0xf6, 0x24, 0x95, 0x53, 0x81, 0xf7, 0xa1, 0x3a, 0x58, 0x23, 0xd7, 0xf7,
	0x0c, 0x54, 0x43, 0x8d, 0xb2, 0x53, 0xd9, 0xb0, 0x8e, 0x87, 0xef, 0x42, 0xc9, 0x1f, 0xd3, 0x21,
	0x8b, 0xe5, 0x8c, 0x92, 0x8b, 0xaa, 0xee, 0x78, 0xf8, 0x01, 0xc0, 0x20, 0x64, 0x54, 0x32, 0xcf,
	0xa5, 0xd2, 0xc8, 0x29, 0xb1, 0x9c, 0x92, 0x96, 0x8c, 0x65, 0x21, 0x69, 0x98, 0xca, 0xf9, 0x44,
	0x4e, 0x49, 0x4b, 0xe2, 0x87, 0x50, 0xf9, 0xe8, 0x07, 0xbe, 0x38, 0x4f, 0xf4, 0x82, 0xd2, 0x61,
	0x8d, 0x5a, 0x12, 0xdf, 0x87, 0x32, 0xfb, 0xe2, 0x4b, 0x77, 0xc0, 0x3d, 0x66, 0x14, 0x6b, 0xa8,
	0x91, 0x77, 0x4a, 0x31, 0x68, 0x73, 0x8f, 0xe1, 0x47, 0xb0, 0x13, 0x32, 0x35, 0xcc, 0x1d, 0xf0,
	0x69, 0x20, 0x8d, 0x92, 0x6a, 0xa8, 0xa6, 0xb0, 0x1d, 0x33, 0x7c, 0x07, 0x0a, 0x21, 0xa3, 0x82,
	0x07, 0x46, 0x45, 0x4d, 0x4f, 0x2b, 0x6c, 0x40, 0x71, 0xcc, 0x84, 0xa0, 0x43, 0x66, 0x54, 0x93,
	0x95, 0xd2, 0xb2, 0xbe, 0x40, 0x50, 0xed, 0x72, 0x2f, 0x89, 0xe7, 0x9d, 0x18, 0x62, 0x1d, 0xb2,
	0xd3, 0x4d, 0x30, 0xf1, 0xcf, 0xf8, 0x71, 0xc0, 0xe4, 0x67, 0x1e, 0x8e, 0x54, 0x1e, 0x55, 0x67,
	0x5d, 0x62, 0x13, 0x60, 0x93, 0x9c, 0x30, 0xb2, 0xb5, 0x6c, 0xa3, 0x72, 0xf8, 0xb8, 0x99, 0xc4,
	0xdf, 0xdc, 0x9e, 0xda, 0xdc, 0xdc, 0x41, 0x58, 0x81, 0x0c, 0x67, 0xce, 0xd6, 0xbb, 0x7b, 0xef,
	0xb7, 0xce, 0x94, 0xc8, 0xb1, 0x89, 0x11, 0x9b, 0xad, 0x4d, 0x8c, 0xd8, 0x0c, 0x1f, 0x40, 0xfe,
	0x82, 0x7e, 0x9a, 0x32, 0x65, 0xa1, 0x72, 0xb8, 0xb7, 0xfe, 0x97, 0x7f, 0x0e, 0xec, 0x24, 0x5d,
	0x2f, 0x32, 0xcf, 0x51, 0xfd, 0x25, 0xe8, 0x1b, 0x0f, 0x47, 0xbe, 0x90, 0xf1, 0x76, 0x0d, 0xc8,
	0x4d, 0xb8, 0x27, 0x0c, 0xa4, 0xbc, 0xee, 0xfe, 0xcf, 0xab, 0xa3, 0x3a, 0x9e, 0x7c, 0x85, 0x52,
	0x4a, 0x19, 0xbe, 0x0d, 0x37, 0xbb, 0xc7, 0xa6, 0xdb, 0x3b, 0x69, 0x9d, 0x58, 0x6e, 0xdf, 0x7e,
	0x6b, 0x1f, 0x9f, 0xda, 0xba, 0x76, 0x1d, 0x77, 0x2d, 0xdb, 0xec, 0xd8, 0x6f, 0x74, 0x74, 0x1d,
	0x3b, 0x7d, 0xdb, 0x8e, 0x71, 0x06, 0xef, 0xc1, 0xad, 0x2b, 0xdc, 0xeb, 0xb7, 0xdb, 0x96, 0x65,
	0x5a, 0xa6, 0x9e, 0xc5, 0xbb, 0xa0, 0x5f, 0x09, 0xaf, 0x5b, 0x9d, 0x23, 0xcb, 0xd4, 0x73, 0xaf,
	0x4e, 0xe7, 0x4b, 0xa2, 0x2d, 0x96, 0x44, 0xbb, 0x5c, 0x12, 0xf4, 0x2d, 0x22, 0xe8, 0x67, 0x44,
	0xd0, 0xaf, 0x88, 0xa0, 0x79, 0x44, 0xd0, 0xef, 0x88, 0xa0, 0x3f, 0x11, 0xd1, 0x2e, 0x23, 0x82,
	0xbe, 0xaf, 0x88, 0x36, 0x5f, 0x11, 0x6d, 0xb1, 0x22, 0xda, 0x87, 0x7d, 0x1a, 0x9e, 0x53, 0xd9,
	0xf4, 0xd8, 0xc5, 0xd3, 0x64, 0xb5, 0x03, 0xf5, 0x4d, 0xa4, 0xc5, 0x90, 0x4f, 0xce, 0xce, 0x0a,
	0x8a, 0x3c, 0xfb, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x48, 0x6b, 0xef, 0x32, 0x36, 0x03, 0x00, 0x00,
}

func (x PodState) String() string {
	s, ok := PodState_name[int32(x)]
	if ok {
		return s
	}
	return strconv.Itoa(int(x))
}
func (this *ContainerStatus) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ContainerStatus)
	if !ok {
		that2, ok := that.(ContainerStatus)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.ContainerId != that1.ContainerId {
		return false
	}
	if this.ImageId != that1.ImageId {
		return false
	}
	if this.CreatedAt != that1.CreatedAt {
		return false
	}
	if this.StartedAt != that1.StartedAt {
		return false
	}
	if this.FinishedAt != that1.FinishedAt {
		return false
	}
	if this.ExitCode != that1.ExitCode {
		return false
	}
	if this.RestartCount != that1.RestartCount {
		return false
	}
	if this.Reason != that1.Reason {
		return false
	}
	if this.Message != that1.Message {
		return false
	}
	return true
}
func (this *PodStatusMsg) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PodStatusMsg)
	if !ok {
		that2, ok := that.(PodStatusMsg)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Uid != that1.Uid {
		return false
	}
	if !bytes.Equal(this.Network, that1.Network) {
		return false
	}
	if len(this.Containers) != len(that1.Containers) {
		return false
	}
	for i := range this.Containers {
		if !this.Containers[i].Equal(that1.Containers[i]) {
			return false
		}
	}
	return true
}
func (this *PodStatusListMsg) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PodStatusListMsg)
	if !ok {
		that2, ok := that.(PodStatusListMsg)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.Pods) != len(that1.Pods) {
		return false
	}
	for i := range this.Pods {
		if !this.Pods[i].Equal(that1.Pods[i]) {
			return false
		}
	}
	return true
}
func (this *ContainerStatus) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 13)
	s = append(s, "&aranyagopb.ContainerStatus{")
	s = append(s, "ContainerId: "+fmt.Sprintf("%#v", this.ContainerId)+",\n")
	s = append(s, "ImageId: "+fmt.Sprintf("%#v", this.ImageId)+",\n")
	s = append(s, "CreatedAt: "+fmt.Sprintf("%#v", this.CreatedAt)+",\n")
	s = append(s, "StartedAt: "+fmt.Sprintf("%#v", this.StartedAt)+",\n")
	s = append(s, "FinishedAt: "+fmt.Sprintf("%#v", this.FinishedAt)+",\n")
	s = append(s, "ExitCode: "+fmt.Sprintf("%#v", this.ExitCode)+",\n")
	s = append(s, "RestartCount: "+fmt.Sprintf("%#v", this.RestartCount)+",\n")
	s = append(s, "Reason: "+fmt.Sprintf("%#v", this.Reason)+",\n")
	s = append(s, "Message: "+fmt.Sprintf("%#v", this.Message)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *PodStatusMsg) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&aranyagopb.PodStatusMsg{")
	s = append(s, "Uid: "+fmt.Sprintf("%#v", this.Uid)+",\n")
	s = append(s, "Network: "+fmt.Sprintf("%#v", this.Network)+",\n")
	keysForContainers := make([]string, 0, len(this.Containers))
	for k, _ := range this.Containers {
		keysForContainers = append(keysForContainers, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForContainers)
	mapStringForContainers := "map[string]*ContainerStatus{"
	for _, k := range keysForContainers {
		mapStringForContainers += fmt.Sprintf("%#v: %#v,", k, this.Containers[k])
	}
	mapStringForContainers += "}"
	if this.Containers != nil {
		s = append(s, "Containers: "+mapStringForContainers+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *PodStatusListMsg) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 5)
	s = append(s, "&aranyagopb.PodStatusListMsg{")
	if this.Pods != nil {
		s = append(s, "Pods: "+fmt.Sprintf("%#v", this.Pods)+",\n")
	}
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringMsgPod(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func (m *ContainerStatus) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ContainerStatus) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ContainerStatus) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Message) > 0 {
		i -= len(m.Message)
		copy(dAtA[i:], m.Message)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.Message)))
		i--
		dAtA[i] = 0x62
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x5a
	}
	if m.RestartCount != 0 {
		i = encodeVarintMsgPod(dAtA, i, uint64(m.RestartCount))
		i--
		dAtA[i] = 0x40
	}
	if m.ExitCode != 0 {
		i = encodeVarintMsgPod(dAtA, i, uint64(m.ExitCode))
		i--
		dAtA[i] = 0x38
	}
	if len(m.FinishedAt) > 0 {
		i -= len(m.FinishedAt)
		copy(dAtA[i:], m.FinishedAt)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.FinishedAt)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.StartedAt) > 0 {
		i -= len(m.StartedAt)
		copy(dAtA[i:], m.StartedAt)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.StartedAt)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.CreatedAt) > 0 {
		i -= len(m.CreatedAt)
		copy(dAtA[i:], m.CreatedAt)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.CreatedAt)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ImageId) > 0 {
		i -= len(m.ImageId)
		copy(dAtA[i:], m.ImageId)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.ImageId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ContainerId) > 0 {
		i -= len(m.ContainerId)
		copy(dAtA[i:], m.ContainerId)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.ContainerId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PodStatusMsg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PodStatusMsg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PodStatusMsg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Containers) > 0 {
		for k := range m.Containers {
			v := m.Containers[k]
			baseI := i
			if v != nil {
				{
					size, err := v.MarshalToSizedBuffer(dAtA[:i])
					if err != nil {
						return 0, err
					}
					i -= size
					i = encodeVarintMsgPod(dAtA, i, uint64(size))
				}
				i--
				dAtA[i] = 0x12
			}
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintMsgPod(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintMsgPod(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Network) > 0 {
		i -= len(m.Network)
		copy(dAtA[i:], m.Network)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.Network)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Uid) > 0 {
		i -= len(m.Uid)
		copy(dAtA[i:], m.Uid)
		i = encodeVarintMsgPod(dAtA, i, uint64(len(m.Uid)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PodStatusListMsg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PodStatusListMsg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PodStatusListMsg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Pods) > 0 {
		for iNdEx := len(m.Pods) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Pods[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMsgPod(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintMsgPod(dAtA []byte, offset int, v uint64) int {
	offset -= sovMsgPod(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ContainerStatus) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContainerId)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.ImageId)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.CreatedAt)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.StartedAt)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.FinishedAt)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	if m.ExitCode != 0 {
		n += 1 + sovMsgPod(uint64(m.ExitCode))
	}
	if m.RestartCount != 0 {
		n += 1 + sovMsgPod(uint64(m.RestartCount))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.Message)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	return n
}

func (m *PodStatusMsg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Uid)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	l = len(m.Network)
	if l > 0 {
		n += 1 + l + sovMsgPod(uint64(l))
	}
	if len(m.Containers) > 0 {
		for k, v := range m.Containers {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.Size()
				l += 1 + sovMsgPod(uint64(l))
			}
			mapEntrySize := 1 + len(k) + sovMsgPod(uint64(len(k))) + l
			n += mapEntrySize + 1 + sovMsgPod(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *PodStatusListMsg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Pods) > 0 {
		for _, e := range m.Pods {
			l = e.Size()
			n += 1 + l + sovMsgPod(uint64(l))
		}
	}
	return n
}

func sovMsgPod(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMsgPod(x uint64) (n int) {
	return sovMsgPod(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *ContainerStatus) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&ContainerStatus{`,
		`ContainerId:` + fmt.Sprintf("%v", this.ContainerId) + `,`,
		`ImageId:` + fmt.Sprintf("%v", this.ImageId) + `,`,
		`CreatedAt:` + fmt.Sprintf("%v", this.CreatedAt) + `,`,
		`StartedAt:` + fmt.Sprintf("%v", this.StartedAt) + `,`,
		`FinishedAt:` + fmt.Sprintf("%v", this.FinishedAt) + `,`,
		`ExitCode:` + fmt.Sprintf("%v", this.ExitCode) + `,`,
		`RestartCount:` + fmt.Sprintf("%v", this.RestartCount) + `,`,
		`Reason:` + fmt.Sprintf("%v", this.Reason) + `,`,
		`Message:` + fmt.Sprintf("%v", this.Message) + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodStatusMsg) String() string {
	if this == nil {
		return "nil"
	}
	keysForContainers := make([]string, 0, len(this.Containers))
	for k, _ := range this.Containers {
		keysForContainers = append(keysForContainers, k)
	}
	github_com_gogo_protobuf_sortkeys.Strings(keysForContainers)
	mapStringForContainers := "map[string]*ContainerStatus{"
	for _, k := range keysForContainers {
		mapStringForContainers += fmt.Sprintf("%v: %v,", k, this.Containers[k])
	}
	mapStringForContainers += "}"
	s := strings.Join([]string{`&PodStatusMsg{`,
		`Uid:` + fmt.Sprintf("%v", this.Uid) + `,`,
		`Network:` + fmt.Sprintf("%v", this.Network) + `,`,
		`Containers:` + mapStringForContainers + `,`,
		`}`,
	}, "")
	return s
}
func (this *PodStatusListMsg) String() string {
	if this == nil {
		return "nil"
	}
	repeatedStringForPods := "[]*PodStatusMsg{"
	for _, f := range this.Pods {
		repeatedStringForPods += strings.Replace(f.String(), "PodStatusMsg", "PodStatusMsg", 1) + ","
	}
	repeatedStringForPods += "}"
	s := strings.Join([]string{`&PodStatusListMsg{`,
		`Pods:` + repeatedStringForPods + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringMsgPod(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *ContainerStatus) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgPod
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ContainerStatus: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ContainerStatus: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContainerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContainerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ImageId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ImageId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedAt", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CreatedAt = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartedAt", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StartedAt = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinishedAt", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FinishedAt = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExitCode", wireType)
			}
			m.ExitCode = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExitCode |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RestartCount", wireType)
			}
			m.RestartCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RestartCount |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 12:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Message", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Message = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgPod(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgPod
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgPod
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
func (m *PodStatusMsg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgPod
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodStatusMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodStatusMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Uid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Uid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Network", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Network = append(m.Network[:0], dAtA[iNdEx:postIndex]...)
			if m.Network == nil {
				m.Network = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Containers", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Containers == nil {
				m.Containers = make(map[string]*ContainerStatus)
			}
			var mapkey string
			var mapvalue *ContainerStatus
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowMsgPod
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMsgPod
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthMsgPod
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthMsgPod
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMsgPod
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthMsgPod
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthMsgPod
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &ContainerStatus{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipMsgPod(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthMsgPod
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Containers[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgPod(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgPod
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgPod
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
func (m *PodStatusListMsg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgPod
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PodStatusListMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodStatusListMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pods", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgPod
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgPod
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Pods = append(m.Pods, &PodStatusMsg{})
			if err := m.Pods[len(m.Pods)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgPod(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgPod
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgPod
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
func skipMsgPod(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMsgPod
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
					return 0, ErrIntOverflowMsgPod
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMsgPod
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
			if length < 0 {
				return 0, ErrInvalidLengthMsgPod
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMsgPod
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMsgPod
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMsgPod        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMsgPod          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMsgPod = fmt.Errorf("proto: unexpected end of group")
)
