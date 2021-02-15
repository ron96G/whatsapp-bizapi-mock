// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.6.1
// source: internal.proto

package model

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

type InternalContact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *InternalContact) Reset() {
	*x = InternalContact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InternalContact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InternalContact) ProtoMessage() {}

func (x *InternalContact) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InternalContact.ProtoReflect.Descriptor instead.
func (*InternalContact) Descriptor() ([]byte, []int) {
	return file_internal_proto_rawDescGZIP(), []int{0}
}

func (x *InternalContact) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *InternalContact) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type InternalConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version      string             `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	Status       string             `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Contacts     []*InternalContact `protobuf:"bytes,3,rep,name=contacts,proto3" json:"contacts,omitempty"`
	UploadDir    string             `protobuf:"bytes,4,opt,name=uploadDir,proto3" json:"uploadDir,omitempty"`
	WebhookUrl   string             `protobuf:"bytes,5,opt,name=webhookUrl,proto3" json:"webhookUrl,omitempty"`
	Users        map[string]string  `protobuf:"bytes,6,rep,name=users,proto3" json:"users,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	InboundMedia map[string]string  `protobuf:"bytes,7,rep,name=inboundMedia,proto3" json:"inboundMedia,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *InternalConfig) Reset() {
	*x = InternalConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InternalConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InternalConfig) ProtoMessage() {}

func (x *InternalConfig) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InternalConfig.ProtoReflect.Descriptor instead.
func (*InternalConfig) Descriptor() ([]byte, []int) {
	return file_internal_proto_rawDescGZIP(), []int{1}
}

func (x *InternalConfig) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *InternalConfig) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *InternalConfig) GetContacts() []*InternalContact {
	if x != nil {
		return x.Contacts
	}
	return nil
}

func (x *InternalConfig) GetUploadDir() string {
	if x != nil {
		return x.UploadDir
	}
	return ""
}

func (x *InternalConfig) GetWebhookUrl() string {
	if x != nil {
		return x.WebhookUrl
	}
	return ""
}

func (x *InternalConfig) GetUsers() map[string]string {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *InternalConfig) GetInboundMedia() map[string]string {
	if x != nil {
		return x.InboundMedia
	}
	return nil
}

type WebhookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Contacts     []*Contact `protobuf:"bytes,1,rep,name=contacts,proto3" json:"contacts,omitempty"`
	Messages     []*Message `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
	Statuses     []*Status  `protobuf:"bytes,3,rep,name=statuses,proto3" json:"statuses,omitempty"`
	Errors       []*Error   `protobuf:"bytes,4,rep,name=errors,proto3" json:"errors,omitempty"`
	ErrorCounter int32      `protobuf:"varint,5,opt,name=errorCounter,proto3" json:"errorCounter,omitempty"`
}

func (x *WebhookRequest) Reset() {
	*x = WebhookRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebhookRequest) ProtoMessage() {}

func (x *WebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebhookRequest.ProtoReflect.Descriptor instead.
func (*WebhookRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_rawDescGZIP(), []int{2}
}

func (x *WebhookRequest) GetContacts() []*Contact {
	if x != nil {
		return x.Contacts
	}
	return nil
}

func (x *WebhookRequest) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

func (x *WebhookRequest) GetStatuses() []*Status {
	if x != nil {
		return x.Statuses
	}
	return nil
}

func (x *WebhookRequest) GetErrors() []*Error {
	if x != nil {
		return x.Errors
	}
	return nil
}

func (x *WebhookRequest) GetErrorCounter() int32 {
	if x != nil {
		return x.ErrorCounter
	}
	return 0
}

var File_internal_proto protoreflect.FileDescriptor

var file_internal_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x1a, 0x0e, 0x77, 0x68, 0x61, 0x74,
	0x73, 0x61, 0x70, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x35, 0x0a, 0x0f, 0x49, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x22, 0xbd, 0x03, 0x0a, 0x0e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x35, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63,
	0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73, 0x12, 0x1c, 0x0a,
	0x09, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x44, 0x69, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x44, 0x69, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x77,
	0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x55, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x55, 0x72, 0x6c, 0x12, 0x39, 0x0a, 0x05, 0x75,
	0x73, 0x65, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x12, 0x4e, 0x0a, 0x0c, 0x69, 0x6e, 0x62, 0x6f, 0x75, 0x6e,
	0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x49, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x4d, 0x65,
	0x64, 0x69, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0c, 0x69, 0x6e, 0x62, 0x6f, 0x75, 0x6e,
	0x64, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x1a, 0x38, 0x0a, 0x0a, 0x55, 0x73, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x1a, 0x3f, 0x0a, 0x11, 0x49, 0x6e, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x4d, 0x65, 0x64, 0x69, 0x61,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0xe9, 0x01, 0x0a, 0x0e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x77, 0x68, 0x61, 0x74, 0x73, 0x61, 0x70,
	0x70, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x63, 0x74, 0x73, 0x12, 0x2d, 0x0a, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x77, 0x68, 0x61, 0x74, 0x73, 0x61, 0x70, 0x70,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x73, 0x12, 0x2c, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x77, 0x68, 0x61, 0x74, 0x73, 0x61, 0x70, 0x70, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73,
	0x12, 0x27, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x77, 0x68, 0x61, 0x74, 0x73, 0x61, 0x70, 0x70, 0x2e, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x52, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x72, 0x42, 0x0a, 0x5a,
	0x08, 0x2e, 0x2e, 0x2f, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_internal_proto_rawDescOnce sync.Once
	file_internal_proto_rawDescData = file_internal_proto_rawDesc
)

func file_internal_proto_rawDescGZIP() []byte {
	file_internal_proto_rawDescOnce.Do(func() {
		file_internal_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_rawDescData)
	})
	return file_internal_proto_rawDescData
}

var file_internal_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_internal_proto_goTypes = []interface{}{
	(*InternalContact)(nil), // 0: internal.InternalContact
	(*InternalConfig)(nil),  // 1: internal.InternalConfig
	(*WebhookRequest)(nil),  // 2: internal.WebhookRequest
	nil,                     // 3: internal.InternalConfig.UsersEntry
	nil,                     // 4: internal.InternalConfig.InboundMediaEntry
	(*Contact)(nil),         // 5: whatsapp.Contact
	(*Message)(nil),         // 6: whatsapp.Message
	(*Status)(nil),          // 7: whatsapp.Status
	(*Error)(nil),           // 8: whatsapp.Error
}
var file_internal_proto_depIdxs = []int32{
	0, // 0: internal.InternalConfig.contacts:type_name -> internal.InternalContact
	3, // 1: internal.InternalConfig.users:type_name -> internal.InternalConfig.UsersEntry
	4, // 2: internal.InternalConfig.inboundMedia:type_name -> internal.InternalConfig.InboundMediaEntry
	5, // 3: internal.WebhookRequest.contacts:type_name -> whatsapp.Contact
	6, // 4: internal.WebhookRequest.messages:type_name -> whatsapp.Message
	7, // 5: internal.WebhookRequest.statuses:type_name -> whatsapp.Status
	8, // 6: internal.WebhookRequest.errors:type_name -> whatsapp.Error
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_internal_proto_init() }
func file_internal_proto_init() {
	if File_internal_proto != nil {
		return
	}
	file_whatsapp_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InternalContact); i {
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
		file_internal_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InternalConfig); i {
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
		file_internal_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebhookRequest); i {
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
			RawDescriptor: file_internal_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_proto_goTypes,
		DependencyIndexes: file_internal_proto_depIdxs,
		MessageInfos:      file_internal_proto_msgTypes,
	}.Build()
	File_internal_proto = out.File
	file_internal_proto_rawDesc = nil
	file_internal_proto_goTypes = nil
	file_internal_proto_depIdxs = nil
}
