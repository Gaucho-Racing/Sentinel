package logger

import "go.uber.org/zap"

var SugarLogger *zap.SugaredLogger

func Init() {
	SugarLogger = zap.Must(zap.NewProduction()).Sugar()
}
