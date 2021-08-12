// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1-devel
// 	protoc        v3.17.3
// source: msg/msg.proto

package msg

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

//Types of messages nodes can send to one another
type Type int32

const (
	Type_Prepare Type = 0
	Type_Promise Type = 1
	Type_Propose Type = 2
	Type_Accept  Type = 3
)

// Enum value maps for Type.
var (
	Type_name = map[int32]string{
		0: "Prepare",
		1: "Promise",
		2: "Propose",
		3: "Accept",
	}
	Type_value = map[string]int32{
		"Prepare": 0,
		"Promise": 1,
		"Propose": 2,
		"Accept":  3,
	}
)

func (x Type) Enum() *Type {
	p := new(Type)
	*p = x
	return p
}

func (x Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type) Descriptor() protoreflect.EnumDescriptor {
	return file_msg_msg_proto_enumTypes[0].Descriptor()
}

func (Type) Type() protoreflect.EnumType {
	return &file_msg_msg_proto_enumTypes[0]
}

func (x Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type.Descriptor instead.
func (Type) EnumDescriptor() ([]byte, []int) {
	return file_msg_msg_proto_rawDescGZIP(), []int{0}
}

//Data type used for inter-node communication
type Msg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type       Type   `protobuf:"varint,1,opt,name=type,proto3,enum=msg.Type" json:"type,omitempty"`
	Id         int64  `protobuf:"varint,2,opt,name=id,proto3" json:"id,omitempty"`
	Value      string `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	PreviousId int64  `protobuf:"varint,4,opt,name=previousId,proto3" json:"previousId,omitempty"` //id of previously accepted value
}

func (x *Msg) Reset() {
	*x = Msg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Msg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Msg) ProtoMessage() {}

func (x *Msg) ProtoReflect() protoreflect.Message {
	mi := &file_msg_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Msg.ProtoReflect.Descriptor instead.
func (*Msg) Descriptor() ([]byte, []int) {
	return file_msg_msg_proto_rawDescGZIP(), []int{0}
}

func (x *Msg) GetType() Type {
	if x != nil {
		return x.Type
	}
	return Type_Prepare
}

func (x *Msg) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Msg) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Msg) GetPreviousId() int64 {
	if x != nil {
		return x.PreviousId
	}
	return 0
}

type MsgSlice struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msgs map[int32]*Msg `protobuf:"bytes,1,rep,name=msgs,proto3" json:"msgs,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MsgSlice) Reset() {
	*x = MsgSlice{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgSlice) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgSlice) ProtoMessage() {}

func (x *MsgSlice) ProtoReflect() protoreflect.Message {
	mi := &file_msg_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgSlice.ProtoReflect.Descriptor instead.
func (*MsgSlice) Descriptor() ([]byte, []int) {
	return file_msg_msg_proto_rawDescGZIP(), []int{1}
}

func (x *MsgSlice) GetMsgs() map[int32]*Msg {
	if x != nil {
		return x.Msgs
	}
	return nil
}

var File_msg_msg_proto protoreflect.FileDescriptor

var file_msg_msg_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6d, 0x73, 0x67, 0x2f, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x6d, 0x73, 0x67, 0x22, 0x6a, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x12, 0x1d, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x09, 0x2e, 0x6d, 0x73, 0x67, 0x2e,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x49, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x49, 0x64,
	0x22, 0x7a, 0x0a, 0x08, 0x4d, 0x73, 0x67, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x12, 0x2b, 0x0a, 0x04,
	0x6d, 0x73, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6d, 0x73, 0x67,
	0x2e, 0x4d, 0x73, 0x67, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x2e, 0x4d, 0x73, 0x67, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x1a, 0x41, 0x0a, 0x09, 0x4d, 0x73, 0x67,
	0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1e, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x73,
	0x67, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x2a, 0x39, 0x0a, 0x04,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x10,
	0x00, 0x12, 0x0b, 0x0a, 0x07, 0x50, 0x72, 0x6f, 0x6d, 0x69, 0x73, 0x65, 0x10, 0x01, 0x12, 0x0b,
	0x0a, 0x07, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x41,
	0x63, 0x63, 0x65, 0x70, 0x74, 0x10, 0x03, 0x32, 0x36, 0x0a, 0x09, 0x4d, 0x65, 0x73, 0x73, 0x65,
	0x6e, 0x67, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x07, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67, 0x12,
	0x0d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x73, 0x67, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x1a, 0x0d,
	0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x73, 0x67, 0x53, 0x6c, 0x69, 0x63, 0x65, 0x22, 0x00, 0x42,
	0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_msg_msg_proto_rawDescOnce sync.Once
	file_msg_msg_proto_rawDescData = file_msg_msg_proto_rawDesc
)

func file_msg_msg_proto_rawDescGZIP() []byte {
	file_msg_msg_proto_rawDescOnce.Do(func() {
		file_msg_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_msg_msg_proto_rawDescData)
	})
	return file_msg_msg_proto_rawDescData
}

var file_msg_msg_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_msg_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_msg_msg_proto_goTypes = []interface{}{
	(Type)(0),        // 0: msg.Type
	(*Msg)(nil),      // 1: msg.Msg
	(*MsgSlice)(nil), // 2: msg.MsgSlice
	nil,              // 3: msg.MsgSlice.MsgsEntry
}
var file_msg_msg_proto_depIdxs = []int32{
	0, // 0: msg.Msg.type:type_name -> msg.Type
	3, // 1: msg.MsgSlice.msgs:type_name -> msg.MsgSlice.MsgsEntry
	1, // 2: msg.MsgSlice.MsgsEntry.value:type_name -> msg.Msg
	2, // 3: msg.Messenger.SendMsg:input_type -> msg.MsgSlice
	2, // 4: msg.Messenger.SendMsg:output_type -> msg.MsgSlice
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_msg_msg_proto_init() }
func file_msg_msg_proto_init() {
	if File_msg_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_msg_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Msg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_msg_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgSlice); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_msg_msg_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_msg_msg_proto_goTypes,
		DependencyIndexes: file_msg_msg_proto_depIdxs,
		EnumInfos:         file_msg_msg_proto_enumTypes,
		MessageInfos:      file_msg_msg_proto_msgTypes,
	}.Build()
	File_msg_msg_proto = out.File
	file_msg_msg_proto_rawDesc = nil
	file_msg_msg_proto_goTypes = nil
	file_msg_msg_proto_depIdxs = nil
}
