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

## What to do when services are upgraded ##

You need to regenerate generated files. Just use gb regenerator in this folder and helpers.go will be updated

```bash
gb generate
```
