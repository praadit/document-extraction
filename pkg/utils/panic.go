package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	constants "textract-mongo/pkg/const"
)

func PanicIfError(err error, logMsg interface{}) {
	if err != nil {

		// ExceptionError(err)
		panic(err)
	}
}

// func ExceptionError(err error) {
// 	if _, ok := err.(*exceptions.MaintenanceError); ok {
// 		panic(err)
// 	}
// 	if _, ok := err.(*exceptions.UnauthorizedError); ok {
// 		panic(err)
// 	}
// 	if _, ok := err.(*exceptions.ForbiddendError); ok {
// 		panic(err)
// 	}
// 	if _, ok := err.(*exceptions.NotFoundError); ok {
// 		panic(err)
// 	}
// 	if _, ok := err.(*exceptions.ValidationError); ok {
// 		panic(err)
// 	}
// 	if _, ok := err.(*exceptions.BadRequestError); ok {
// 		panic(err)
// 	}
// }

// func throwPanicPgError(code string, mesage string) {
// 	panic(&exceptions.InternalError{
// 		Code:            code,
// 		InternalMessage: mesage,
// 	})
// }

func getCaller(skip int) (string, int) {
	_, path, line, ok := runtime.Caller(skip)
	filename := ""
	if ok {
		filename = filepath.Base(path)
	} else {
		line = 0
	}
	return filename, line
}

func SqlPanicFilter(ctx context.Context, err error, message, notfoundMsg string) (bool, error) {
	if err == nil {
		return false, nil
	}
	id := ""
	if ctx != nil {
		id = getLogIdentifier(ctx)
	}

	filename, line := getCaller(2)

	switch err {
	case nil:
		return false, nil
	case sql.ErrNoRows:
		if len(notfoundMsg) < 1 {
			// log.Printf("%s - [%s:%d] %s - error : %s\n", id, filename, line, "not found", err.Error())
			return true, fmt.Errorf("record not found")
		} else {
			// log.Printf("%s - [%s:%d] %s - error : %s\n", id, filename, line, notfoundMsg, err.Error())
			return true, fmt.Errorf(notfoundMsg)
		}
	default:
		log.Printf("%s - [%s:%d] %s - error : %s\n", id, filename, line, message, err.Error())
		return false, fmt.Errorf(message)
	}
}

func getLogIdentifier(ctx context.Context) string {
	logId := ctx.Value(constants.LOGGERID)
	if logId != nil {
		return logId.(string)
	}
	return ""
}

func FilterError(ctx context.Context, err error, message string) error {
	if err == nil {
		return nil
	}

	id := ""
	if ctx != nil {
		id = getLogIdentifier(ctx)
	}

	filename, line := getCaller(2)

	log.Printf("%s - [%s:%d] %s - error : %s\n", id, filename, line, message, err.Error())
	return fmt.Errorf(message)
}

func Log(ctx context.Context, message string) {
	id := ""
	if ctx != nil {
		id = getLogIdentifier(ctx)
	}

	log.Printf("[%s] %s\n", id, message)
}
