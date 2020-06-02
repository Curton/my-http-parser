package parser

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

// 混合数据, 两次缓冲区读取
func TestTwoMix(t *testing.T) {
	//未知数据 + 一个复杂GET请求 + 一个简单GET请求 + 一个简单POST请求 + 无效body + 一个简单GET请求 + 无效数据 + 半个 POST
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
		"\r\nGET /path2?ie=utf-8&mod=1&_cr1=1234 HTTP/1.1\r\n\r\nPOST /path3?_cr1=12345 HTTP/1.7\r\nHost: www.qq.com\r\nContent-Length: 4\r\n\r\nOKOKgvhjsdfhjvgjvGET /path4 HTTP/1.3\r\nLanguage: utf-8\r\n\r\njggjvPOST "

	rq1 := []byte(requestStr1)

	// 另一段缓冲数据: 上一段剩余的POST (no body) + POST
	requestStr2 := " /path5?ie=utf-8&mod=1&_cr1=99999 HTTP/1.1\r\n\r\nPOST /path6?_cr1=12345 HTTP/1.1\r\nHost: www.qq.com\r\nContent-Length: 4\r\n\r\nhaha"

	rq2 := []byte(requestStr2)

	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest, 100)
	go HttpRequestHandler(bufCh, resultCh)

	wg := sync.WaitGroup{}
	wg.Add(6)

	go func() {
		bufCh <- &rq1
		bufCh <- &rq2
	}()

	var r *HttpRequest
	go func() {
		for {
			r = <-resultCh
			fmt.Printf("Method : %s \r\n", r.Method)
			fmt.Printf("Path : %s \r\n", r.Path)
			fmt.Printf("Queries : %s \r\n", r.Queries)
			fmt.Printf("Version : %s \r\n", r.Version)
			fmt.Printf("Headers : %s \r\n", r.Headers)
			fmt.Printf("Body : %s \r\n", r.Body)
			fmt.Println()
			wg.Done()
		}
	}()
	wg.Wait()
}

// 超长POST
func TestLongBody(t *testing.T) {
	// 超长 body 请求, len > 1024, 主动放弃解析body部分, 保留header 用于后续错误处理
	// 1. Content-Length 超长
	requestStr1 := "opiqergifrnbhiPOST /?ie=utf-8&mod=1&_cr1=38721 HTTP/1.1\r\nHost: www.kittyhell.com\r\nContent-Length: 2500\r\n\r\nfahsjvjhsdvfjhvdjvfhjsdhjavfhjvjdhsvasvhjcjhvjhvj"
	rq1 := []byte(requestStr1)
	// Content-Length 超长

	// 2. 实际数据超过 Content-Length, 只保留声明部分内容
	requestStr2 := "osdbnjklrtnnbovcflgiPOST /hi HTTP/1.2\r\nHost: www.baidu.com\r\nContent-Length: 4\r\n\r\nOKfahsjvjhsdvfjhvdjvfhjsdhjavfhjvjdhsvasvhjcjhvjhvjl"
	rq2 := []byte(requestStr2)

	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)

	wg := sync.WaitGroup{}
	wg.Add(16)
	go func() {
		bufCh <- &rq1
		bufCh <- &rq2
		bufCh <- &rq1
		bufCh <- &rq2

		bufCh <- &rq2
		bufCh <- &rq1
		bufCh <- &rq2
		bufCh <- &rq1

		bufCh <- &rq1
		bufCh <- &rq1
		bufCh <- &rq1
		bufCh <- &rq1

		bufCh <- &rq2
		bufCh <- &rq2
		bufCh <- &rq2
		bufCh <- &rq2
	}()

	go func() {
		for {
			r := <-resultCh
			fmt.Printf("Method : %s \r\n", r.Method)
			fmt.Printf("Path : %s \r\n", r.Path)
			fmt.Printf("Queries : %s \r\n", r.Queries)
			fmt.Printf("Version : %s \r\n", r.Version)
			fmt.Printf("Headers : %s \r\n", r.Headers)
			fmt.Printf("Body : %s \r\n", r.Body)
			fmt.Println()
			wg.Done()
		}
	}()
	wg.Wait()
}

// wait POST BODY, 3 pieces
func TestWaitPostBody(t *testing.T) {
	// 1. Content-Length 不足够的POST BODY
	requestStr1 := "POST /?ie=utf-8&mod=1&_cr1=38721 HTTP/1.1\r\nHost: www.kittyhell.com\r\nContent-Length: 6\r\n\r\nOK"
	rq1 := []byte(requestStr1)
	// Content-Length 超长

	// 2. 中间的的 POST BODY
	requestStr2 := "AB"
	rq2 := []byte(requestStr2)

	// 3. 剩余的 POST BODY
	requestStr3 := "CD"
	rq3 := []byte(requestStr3)

	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		bufCh <- &rq1
		bufCh <- &rq2
		bufCh <- &rq3
	}()

	go func() {
		for {
			r := <-resultCh
			fmt.Printf("Method : %s \r\n", r.Method)
			fmt.Printf("Path : %s \r\n", r.Path)
			fmt.Printf("Queries : %s \r\n", r.Queries)
			fmt.Printf("Version : %s \r\n", r.Version)
			fmt.Printf("Headers : %s \r\n", r.Headers)
			fmt.Printf("Body : %s \r\n", r.Body)
			wg.Done()
		}
	}()
	wg.Wait()
}

var test01 = []byte("GET / HTTP/1.1\r\n\r\n")

// 最简单GET性能测试
func BenchmarkHttpRequestHandler01(b *testing.B) {
	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)
	go func() {
		for {
			<-resultCh
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufCh <- &test01
	}
}

var test02 = []byte("GET / HTTP/1.1\r\nHost: google.com\r\nCookie: NID=204=Skq9Nm39xa5b1DYfg7rXScZuFKXrQP1bPL-FftiY2juL9t05Oc_QvWAKOi20QY_o5f0nT7--eIGAAydMOS_ia03vlSMRgIahwuMK9SoLEXZ88yrLTRotctUoFuWdCYBWzBLls6BFv1jb1WVlffskxJ597M5iuT6hqgZVVTX23-k\r\n\r\n")

// 基本GET性能测试
func BenchmarkHttpRequestHandler02(b *testing.B) {
	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)
	go func() {
		for {
			<-resultCh
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufCh <- &test02
	}
}

var test03 = []byte("POST /?ie=utf-8&mod=1&_cr1=12345 HTTP/1.1\r\nHost: www.fucking.com\r\nContent-Length: 4\r\n\r\nOKOK")

// POST性能测试
func BenchmarkHttpRequestHandler03(b *testing.B) {
	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)
	go func() {
		for {
			<-resultCh
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufCh <- &test03
	}
}

var test04 = []byte("GET /?q1=1&q2=2&q3=3&q4=4&q5=5&q6=6&q7=7&q8=8&q9=9&q10=10&q11=11 HTTP/1.1\r\n\r\n")

// long queries look up: 11 queries search by map
// BenchmarkHttpRequestHandler04-4   	 2857335	       410 ns/op
func BenchmarkHttpRequestHandler04(b *testing.B) {
	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)

	bufCh <- &test04
	r := <-resultCh

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i := 1; i < 12; i++ {
			r.FindQuery("q" + strconv.Itoa(i))
		}
	}

}

var test05 = []byte("GET /?q1=1&q2=2&q3=3&q4=4&q5=5&q6=6&q7=7&q8=8&q9=9&q10=10 HTTP/1.1\r\n\r\n")

// long queries look up: 10 queries search by for loop (backend struct will set up a map only when size > 10)
// BenchmarkHttpRequestHandler05-4   	 1205402	       977 ns/op
func BenchmarkHttpRequestHandler05(b *testing.B) {
	bufCh := make(chan *[]byte)
	resultCh := make(chan *HttpRequest)
	go HttpRequestHandler(bufCh, resultCh)

	bufCh <- &test05
	r := <-resultCh

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for i := 1; i < 11; i++ {
			r.FindQuery("q" + strconv.Itoa(i))
		}
	}
}
