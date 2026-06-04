package logger

import (
	"io"
	"log"
	"os"
)

func InitLogger() {
	file, err := os.OpenFile(
		"logs/app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		log.Fatal(err)
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
