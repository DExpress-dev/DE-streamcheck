package main

import (
	"fmt"
	"product_code/check_stream/config"
	"strings"
	//	"product_code/check_stream/public"
	"testing"
	"time"
)

func Test(t *testing.T) {
	config.GetInstance()
	//	fmt.Println(public.TimeFromString("2019-1-22 15:30:00"))

	tsName := "309572926.ts"
	fmt.Println(strings.TrimSuffix(tsName, ".ts"))
	fmt.Println(time.ParseInLocation("2006-1-2 15:04:05", "2025-05-03 15:30:00", time.Local))
}
