package log

import "github.com/getsentry/raven-go"


func ravenErrorLogger(err error, tags map[string]string){
	packet := raven.NewPacket(err.Error(), raven.NewException(err, raven.NewStacktrace(3, 3, nil)))
	level, _ := tags["level"]
	packet.Level = raven.Severity(tags["level"])
	delete(tags,"level")
	_, ch := ravenClient.Capture(packet, tags)
	if level == LevelPanic || level == LevelFatal {
		<-ch
	}
}

func ravenMessageLogger(msg string, tags map[string]string){
	//don't log debug messages to sentry
	if tags["level"] == LevelDebug {
		return
	}
	packet := raven.NewPacket(msg)
	packet.Level = raven.Severity(tags["level"])
	delete(tags,"level")
	ravenClient.Capture(packet, tags)
}
