## protocall ##

Simple grpc test tool.

Input format is yaml.

Usage example:

```bash
protocall localhost:4040 TelegramService NotifyMessage << EOF 
channel: new_user
message: cococotest
EOF
```

