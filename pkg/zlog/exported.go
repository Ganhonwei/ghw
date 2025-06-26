package zlog

func Info(args ...any) {
	if Logger != nil {
		Logger.Info(args...)
	}
}
func Infoln(args ...any) {
	if Logger != nil {
		Logger.Infoln(args...)
	}
}
func Infof(format string, args ...any) {
	if Logger != nil {
		Logger.Infof(format, args...)
	}
}
func Warning(args ...any) {
	if Logger != nil {
		Logger.Warn(args...)
	}
}
func Warningln(args ...any) {
	if Logger != nil {
		Logger.Warnln(args...)
	}
}
func Warningf(format string, args ...any) {
	if Logger != nil {
		Logger.Warnf(format, args...)
	}
}
func Error(args ...any) {
	if Logger != nil {
		Logger.Error(args...)
	}
}
func Errorln(args ...any) {
	if Logger != nil {
		Logger.Errorln(args...)
	}
}
func Errorf(format string, args ...any) {
	if Logger != nil {
		Logger.Errorf(format, args...)
	}
}
func Fatal(args ...any) {
	if Logger != nil {
		Logger.Fatal(args...)
	}
}
func Fatalln(args ...any) {
	if Logger != nil {
		Logger.Fatalln(args...)
	}
}
func Fatalf(format string, args ...any) {
	if Logger != nil {
		Logger.Fatalf(format, args...)
	}
}
