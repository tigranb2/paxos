package logger

import (
	"log"
	"os"
)

func WriteToLog(file, str string) {
	f, err := os.OpenFile("logs/"+file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	if _, err = f.WriteString(str); err != nil {
		panic(err)
	}
}
