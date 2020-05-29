package main

import (
	"fmt"
	parser "my-http-parser"
)

func main() {
	// 第一段请求
	//未知数据 + 一个复杂GET请求 + 一个简单GET请求 + 一个简单POST请求
	requestStr1 := "jkdafbkjvbjkdabfkvbsdfghedfhbfhdjGET /wp-content/uploads/2010/03/hello-kitty-darth-vader-pink.jpg HTTP/1.1\r\n" +
		"Host: www.kittyhell.com\r\n" +
		"User-Agent: Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; ja-JP-mac; rv:1.9.2.3) Gecko/20100401 Firefox/3.6.3 " +
		"Pathtraq/0.9\r\n" +
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
		"Accept-Language: ja,en-us;q=0.7,en;q=0.3\r\n" +
		"Accept-Encoding: gzip,deflate\r\n" +
		"Accept-Charset: Shift_JIS,utf-8;q=0.7,*;q=0.7\r\n" +
		"Keep-Alive: 115\r\n" +
		"Connection: keep-alive\r\n" +
		"Cookie: wp_ozh_wsa_visits=2; wp_ozh_wsa_visit_lasttime=xxxxxxxxxx; " +
		"__utma=xxxxxxxxx.xxxxxxxxxx.xxxxxxxxxx.xxxxxxxxxx.xxxxxxxxxx.x; " +
		"__utmz=xxxxxxxxx.xxxxxxxxxx.x.x.utmccn=(referral)|utmcsr=reader.livedoor.com|utmcct=/reader/|utmcmd=referral\r\n" +
		"\r\nGET /?ie=utf-8&mod=1&_cr1=38721 HTTP/1.1\r\n\r\nPOST /?ie=utf-8&mod=1&_cr1=38721 HTTP/1.1\r\nHost: www.kittyhell.com\r\nContent-Length: 4\r\n\r\nOKKgvgjvGET / HTTP/1.1\r\nLanguage: utf-8\r\n\r\njggjv"

	rq1 := []byte(requestStr1)

	// http 解析

	//processed := 0 // 处理的字节数
	//lastRet := 0

	//for processed != len(rq1) {
	//	if lastRet < 0 {
	//		break
	//	}
	//
	//	var method []byte
	//	var path []byte
	//	var minorVersion int
	//	headers := make([]parser.Header, 0, 1<<5)
	//	queries := make([]parser.Query, 0, 1<<5)
	//
	//	r := rq1[processed:]
	//	lastRet = parser.ParseRequest(&r, &method, &path, &queries, &minorVersion, &headers)
	//	processed += lastRet
	//	fmt.Printf("processed : %v \r\n", processed)
	//	fmt.Printf("method : %s \r\n", method)
	//	fmt.Printf("path : %s \r\n", path)
	//	fmt.Printf("queries : %s \r\n", queries)
	//	fmt.Printf("version : 1.%v \r\n", minorVersion)
	//	fmt.Printf("headers : %s \r\n", headers)
	//}
	bufCh := make(chan *[]byte)
	resultCh := make(chan *parser.HttpRequest)
	go parser.HttpRequestHandler(bufCh, resultCh)

	go func() {
		bufCh <- &rq1
	}()

	var r *parser.HttpRequest
	for {
		select {
		case r = <-resultCh:
			fmt.Printf("Method : %s \r\n", r.Method)
			fmt.Printf("Path : %s \r\n", r.Path)
			fmt.Printf("Queries : %s \r\n", r.Queries)
			fmt.Printf("Headers : %s \r\n", r.Headers)
			fmt.Printf("Body : %s \r\n", r.Body)
		default:
			// do nothing
		}
	}
}
