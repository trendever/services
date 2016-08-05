# proto
Proto files for protobuf

After any changes you shoud regenerate all the .pb.go files with ``gb generate``

# Style Guide

Use Request/Reply words for messages to mark them as request or reply from the server

Example:

```
//This is request message
message HelloRequest {
  string name = 1;
}
//This is reply message
message HelloReply {
  string message = 1;
}
```
