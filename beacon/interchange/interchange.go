package interchange

//go:generate protoc --proto_path=./ -I./ --go_out=./ device_message.proto
//go:generate protoc --proto_path=./ -I./ --go_out=./ control_message.proto
//go:generate protoc --proto_path=./ -I./ --go_out=./ welcome_message.proto
//go:generate protoc --proto_path=./ -I./ --go_out=./ feedback_message.proto
//go:generate protoc --proto_path=./ -I./ --go_out=./ error_message.proto
//go:generate protoc --proto_path=./ -I./ --go_out=./ report_message.proto
