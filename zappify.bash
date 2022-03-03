#!/usr/bin/env bash

perl -pi -e "s/log.Debug/log.Info/g" `mdfind -onlyin . -name "*.go" | grep -v vendor`
perl -pi -e "s/log.Trace/log.Debug/g" `mdfind -onlyin . -name "*.go" | grep -v vendor`
perl -pi -e "s/golog.LoggerFor/zaplog.LoggerFor/g" `mdfind -onlyin . -name ".go" | grep -v vendor`
perl -pi -e 's;"github.com/getlantern/golog";"github.com/getlantern/zaplog";g' `mdfind -onlyin . -name ".go" | grep -v vendor`

