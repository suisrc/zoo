package log

import (
	logfile "github.com/suisrc/zoo/zoe/log/file"
	logsyslog "github.com/suisrc/zoo/zoe/log/syslog"
)

var (
	NewFileWriter   = logfile.NewWriter
	NewSyslogWriter = logsyslog.NewWriter
)
