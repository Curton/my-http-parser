/*
 * Covey Liu, covey@liukedun.com  2020/5/11 18:40
 *
 * This parser implementation is reference to https://github.com/h2o/picohttpparser written in C
 *
 */

package parser

import (
	"bytes"
	"regexp"
	"strconv"
	"sync"
)

type Header struct {
	name  []byte
	value []byte
}

type Query struct {
	key   []byte
	value []byte
}

//var httpMethodPattern, _ = regexp.Compile("GET |HEAD |POST |PUT |DELETE |CONNECT |OPTIONS |TRACE |PATCH ")
var httpMethodPattern, _ = regexp.Compile("GET |POST ")
var tokenCharMap = []byte("\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\001\000\001\001\001\001\001\000\000\001\001\000\001\001\000\001\001\001\001\001\001\001\001\001\001\000\000\000\000\000\000\000\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\000\000\000\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\001\000\001\000\001\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000")

func parseHttpVersion(buf *[]byte, minorVersion *int, ret *int) *[]byte {
	bufIndex := 0
	bufEndIndex := len(*buf)

	if bufEndIndex < 9 {
		*ret = -2
		return nil
	}
	i := bufIndex
	bufIndex++
	if (*buf)[i] != 'H' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != 'T' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != 'T' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != 'P' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != '/' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != '1' {
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	if (*buf)[i] != '.' {
		*ret = -1
		return nil
	}
	if (*buf)[bufIndex] < '0' || '9' < (*buf)[bufIndex] {
		bufIndex++
		*ret = -1
		return nil
	}
	i = bufIndex
	bufIndex++
	*minorVersion = int((*buf)[i] - '0')

	tmp := (*buf)[bufIndex:]
	return &tmp
}

func parseRequest(buf *[]byte, method *[]byte, path *[]byte, minorVersion *int, headers *[]Header, ret *int) *[]byte {

	/* skip first empty line (some clients add CRLF after POST content) */
	bufIndex := 0
	bufEndIndex := len(*buf)
	// check EOF
	if bufIndex == bufEndIndex {
		*ret = -2
		return nil
	}

	if (*buf)[bufIndex] == '\r' {
		bufIndex++
		// check EOF
		if bufIndex == bufEndIndex {
			*ret = -2
			return nil
		}
		// expect char
		if (*buf)[bufIndex] == '\n' {
			*ret = -1
			return nil
		}
	} else if (*buf)[bufIndex] == '\n' {
		bufIndex++
	}

	/* parse request line */
	// search METHOD
	{
		tokStart := bufIndex

		for bufIndex < bufEndIndex {
			if (*buf)[bufIndex] == ' ' {
				break
			} else if (*buf)[bufIndex] < '\040' || (*buf)[bufIndex] == '\177' {
				*ret = -1
				return nil
			}
			bufIndex++
			if bufIndex == bufEndIndex {
				*ret = -2
				return nil
			}
		}
		*method = (*buf)[tokStart:bufIndex]
	}

	// do {} while(0)
	{
		bufIndex++
		for bufIndex < bufEndIndex {
			if (*buf)[bufIndex] != ' ' {
				break
			}
			bufIndex++
		}
	}

	// search PATH
	{
		tokStart := bufIndex

		for bufIndex < bufEndIndex {
			if (*buf)[bufIndex] == ' ' {
				break
			} else if (*buf)[bufIndex] < '\040' || (*buf)[bufIndex] == '\177' {
				*ret = -1
				return nil
			}
			bufIndex++
			if bufIndex == bufEndIndex {
				*ret = -2
				return nil
			}
		}
		*path = (*buf)[tokStart:bufIndex]
	}

	// do {} while(0)
	{
		bufIndex++
		for bufIndex < bufEndIndex {
			if (*buf)[bufIndex] != ' ' {
				break
			}
			bufIndex++
		}
	}

	if len(*method) == 0 || len(*path) == 0 {
		*ret = -1
		return nil
	}

	tmp := (*buf)[bufIndex:]
	buf = parseHttpVersion(&tmp, minorVersion, ret)
	if buf == nil {
		return nil
	}

	bufIndex = 0
	bufEndIndex = len(*buf)
	if (*buf)[bufIndex] == '\015' {
		bufIndex++
		if bufIndex == bufEndIndex {
			*ret = -2
			return nil
		}

		i := bufIndex
		bufIndex++
		if (*buf)[i] != '\012' {
			*ret = -1
			return nil
		}
	} else if (*buf)[bufIndex] == '\012' {
		bufIndex++
	} else {
		*ret = -1
		return nil
	}
	tmp2 := (*buf)[bufIndex:]
	return parseHeaders(&tmp2, headers, ret)
}

func parseHeaders(buf *[]byte, headers *[]Header, ret *int) *[]byte {

	bufIndex := 0
	bufEndIndex := len(*buf)
	//headerLen := len(*Headers)
	headerIndex := 0
	maxHeaders := cap(*headers)
	for ; headerIndex < maxHeaders; headerIndex++ {
		if bufIndex == bufEndIndex {
			*ret = -2
			return nil
		}
		if (*buf)[bufIndex] == '\015' {
			bufIndex++
			if bufIndex == bufEndIndex {
				*ret = -2
				return nil
			}

			i := bufIndex
			bufIndex++
			if (*buf)[i] != '\012' {
				*ret = -1
				return nil
			}
			break
		} else if (*buf)[bufIndex] == '\012' {
			bufIndex++
			break
		}
		if !(headerIndex != 0 && ((*buf)[bufIndex] == ' ' || (*buf)[bufIndex] == '\t')) {
			tmp := (*buf)[bufIndex:]
			emptyHeader := Header{}
			appended := append(*headers, emptyHeader)
			*headers = appended
			(*headers)[headerIndex].name = tmp

			for {
				if (*buf)[bufIndex] == ':' {
					break
				} else if tokenCharMap[(*buf)[bufIndex]] == 0 {
					*ret = -1
					return nil
				}
				bufIndex++
				if bufIndex == bufEndIndex {
					*ret = -2
					return nil
				}
			}

			name := (*headers)[headerIndex].name
			nameLen := len(name)
			nameEnd := bufIndex - (len(*buf) - nameLen)
			(*headers)[headerIndex].name = name[:nameEnd]

			if nameLen == nameEnd {
				*ret = -1
				return nil
			}
			bufIndex++

			for ; ; bufIndex++ {
				if bufIndex == bufEndIndex {
					*ret = -2
					return nil
				}
				if !((*buf)[bufIndex] == ' ' || (*buf)[bufIndex] == '\t') {
					break
				}
			}
		} else {
			(*headers)[headerIndex].name = nil
		}

		var value []byte
		var buf2 *[]byte
		tmp := (*buf)[bufIndex:]
		buf2 = getTokenToEol(&tmp, &value, ret)
		bufIndex += (bufEndIndex - bufIndex) - len(*buf2)
		if buf2 == nil {
			return nil
		}

		valueEndIndex := len(value) - 1
		for ; valueEndIndex != 0; valueEndIndex-- {
			c := value[valueEndIndex]
			if !(c == ' ' || c == '\t') {
				break
			}
		}

		(*headers)[headerIndex].value = value[:valueEndIndex+1]

	}

	tmp := (*buf)[bufIndex:]
	return &tmp
}

func isPrintableASCII(b byte) bool {
	return b-040 < 0137
}

func getTokenToEol(buf *[]byte, token *[]byte, ret *int) *[]byte {

	bufIndex := 0
	bufEndIndex := len(*buf)

	for bufEndIndex-bufIndex >= 8 {
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		{
			if !isPrintableASCII((*buf)[bufIndex]) {
				goto NonPrintable
			}
			bufIndex++
		}
		continue

	NonPrintable:
		b := (*buf)[bufIndex]
		if b < '\040' && b != '\011' || b == '\177' {
			goto FoundCtl
		}
		bufIndex++
	}

	for ; ; bufIndex++ {
		// check EOF
		if bufIndex == bufEndIndex {
			*ret = -2
			return nil
		}
		if !((*buf)[bufIndex]-040 < 0137) {
			if (*buf)[bufIndex] < '\040' && (*buf)[bufIndex] != '\011' || (*buf)[bufIndex] == '\177' {
				goto FoundCtl
			}
		}
	}

FoundCtl:

	if (*buf)[bufIndex] == '\015' {
		bufIndex++
		// check EOF
		if bufIndex == bufEndIndex {
			*ret = -2
			return nil
		}
		i := bufIndex
		bufIndex++
		if (*buf)[i] != '\012' {
			*ret = -1
			return nil
		}
		// token_len
		*token = (*buf)[:bufIndex-2]
	} else if (*buf)[bufIndex] == '\012' {
		*buf = (*buf)[:bufIndex]
		bufIndex++
	} else {
		*ret = -1
		return nil
	}

	//*token = (*token)[tokenStart:]
	tmp := (*buf)[bufIndex:]
	return &tmp
}

func ParseRequest(buf *[]byte, method *[]byte, path *[]byte, queries *[]Query, minorVersion *int, headers *[]Header) int {

	buf2 := buf
	ret := 0

	*method = nil
	*path = nil
	*minorVersion = -1

	// try to find first http Method position
	buf2 = findHttpMethodIndex(buf2)
	if buf2 == nil {
		return -2
	}

	buf2 = parseRequest(buf2, method, path, minorVersion, headers, &ret)
	if buf2 == nil {
		return ret
	}
	parseQuery(path, queries)
	return len(*buf) - len(*buf2)
}

func findHttpMethodIndex(buf *[]byte) *[]byte {
	index := httpMethodPattern.FindIndex(*buf)
	if index == nil {
		return nil
	}
	tmp := (*buf)[index[0]:]
	return &tmp
}

func parseQuery(path *[]byte, queries *[]Query) {
	pathLen := len(*path)
	indexByte := bytes.IndexByte(*path, '?')
	if indexByte == -1 || indexByte == pathLen-1 /* only a '?' */ {
		return
	}
	indexByte++
	query := (*path)[indexByte:]
	*path = (*path)[:indexByte-1]
	queryIndex := 0
	queryLen := len(query)

	keyStart := 0
	valueStart := 0
	settingKey := true
	settingValue := false

	var currentQuery *Query
	for queryIndex < queryLen {
		c := query[queryIndex]
		if settingKey && (c == '=' || c == '&') {
			tmp := Query{}
			tmp.key = query[keyStart:queryIndex]
			ap := append(*queries, tmp)
			currentQuery = &ap[len(ap)-1]
			*queries = ap
			// finish setting key
			settingKey = false
			settingValue = true
			valueStart = queryIndex + 1
		}

		if settingValue && c == '&' {
			(*currentQuery).value = query[valueStart:queryIndex]
			// finish setting value
			settingKey = true
			settingValue = false
			keyStart = queryIndex + 1
		}

		if queryIndex == queryLen-1 {
			(*currentQuery).value = query[valueStart:]
			return
		}
		queryIndex++
	}

}

type HttpRequest struct {
	Method  []byte
	Path    []byte
	Queries []Query
	Version []byte
	Headers []Header
	Body    []byte

	// will be initialize only the size > 10, for quick look up
	queriesMap map[string][]byte
	headersMap map[string][]byte
}

func (r *HttpRequest) FindQuery(str string) []byte {
	if r.queriesMap != nil {
		return r.queriesMap[str]
	}
	for _, v := range r.Queries {
		if bytes.Equal(v.key, []byte(str)) {
			return v.value
		}
	}
	return nil
}

func (r *HttpRequest) FindHeader(str string) []byte {
	if r.headersMap != nil {
		return r.headersMap[str]
	}
	for _, v := range r.Headers {
		if bytes.Equal(v.name, []byte(str)) {
			return v.value
		}
	}
	return nil
}

// using map while size > 16 for quick look up
func (r *HttpRequest) init() *HttpRequest {
	qLen := len(r.Queries)
	hLen := len(r.Headers)
	if qLen > 10 {
		r.queriesMap = make(map[string][]byte, qLen)
		for i := 0; i < qLen; i++ {
			r.queriesMap[string(r.Queries[i].key)] = r.Queries[i].value
		}
	}
	if hLen > 10 {
		r.headersMap = make(map[string][]byte, hLen)
		for i := 0; i < hLen; i++ {
			r.headersMap[string(r.Headers[i].name)] = r.Headers[i].value
		}
	}

	return r
}

// should run in goroutine
func HttpRequestHandler(bufCh chan *[]byte, resultCh chan *HttpRequest) {
	waitPostBody := false
	var waitToBodyLen uint64
	var waitPostBodyHttpRequest HttpRequest
	lastRemainBuf := bytes.Buffer{}
	lastInvalid := false

	for {
		// read new data from channel
		var buf *[]byte

		// check if read new data from bufCh
		if lastRemainBuf.Len() == 0 || lastInvalid {
			buf = <-bufCh
			if lastRemainBuf.Len() != 0 {
				// attach with last remain
				tmp := lastRemainBuf.Bytes()
				tmp = append(tmp, *buf...)
				buf = &tmp
			}
			lastInvalid = false
		} else if waitPostBody {
			buf = <-bufCh
			// attach with last remain
			tmp := lastRemainBuf.Bytes()
			tmp = append(tmp, *buf...)
			buf = &tmp
		} else {
			// process last remain
			tmp := lastRemainBuf.Bytes()
			buf = &tmp
		}

		if waitPostBody {
			if uint64(len(*buf)) >= waitToBodyLen {
				waitPostBodyHttpRequest.Body = (*buf)[:waitToBodyLen]
				resultCh <- waitPostBodyHttpRequest.init()
				waitPostBody = false
				lastRemainBuf.Reset()
				lastRemainBuf.Write((*buf)[waitToBodyLen:])
				continue
			} else {
				// continue to wait for data
				// write back data to buf
				lastRemainBuf.Reset()
				lastRemainBuf.Write(*buf)
				continue
			}
		}

		// new request
		method := make([]byte, 0, 1<<2)
		path := make([]byte, 0, 1<<3)
		var minorVersion int
		headers := make([]Header, 0, 1<<5)
		queries := make([]Query, 0, 1<<5)
		//request := HttpRequest{}

		buf2 := make([]byte, len(*buf))
		copy(buf2, *buf)
		processed := ParseRequest(&buf2, &method, &path, &queries, &minorVersion, &headers)

		if processed > 0 {

			// GET request
			if bytes.Equal(method, []byte("GET")) {
				ver := "HTTP/1." + strconv.Itoa(minorVersion)

				resultCh <- (&HttpRequest{
					Method:  method,
					Path:    path,
					Queries: queries,
					Version: []byte(ver),
					Headers: headers,
					Body:    nil,
				}).init()
				if len(*buf) == processed {
					lastRemainBuf.Reset()
				} else {
					lastRemainBuf.Reset()
					lastRemainBuf.Write((*buf)[processed:])
				}
				lastInvalid = false
			}
			// POST request
			if bytes.Equal(method, []byte("POST")) {
				ver := "HTTP/1." + strconv.Itoa(minorVersion)
				var postHttpRequest = HttpRequest{
					Method:  method,
					Path:    path,
					Queries: queries,
					Version: []byte(ver),
					Headers: headers,
					Body:    nil,
				}
				// get Content-Length
				isInvalidPost := false
				waitToBodyLen = 0
				for _, q := range headers {
					if bytes.Equal(q.name, []byte("Content-Length")) {
						contentLen, err := strconv.ParseUint(string(q.value), 10, 64)
						// err in process Content-Length or ~ == 0
						if err != nil || contentLen == 0 || contentLen > 1<<10 {
							isInvalidPost = true
							lastInvalid = true
							//waitToBodyLen = 0
							break
						}
						waitToBodyLen = contentLen
						break
					}
				}

				// no declared Content-Length or invalid content length
				if waitToBodyLen == 0 {
					resultCh <- postHttpRequest.init()
					lastRemainBuf.Reset()
					lastRemainBuf.Write((*buf)[processed:])
				} else if isInvalidPost {
					// declared Content-Length too long
					lastRemainBuf.Reset()
					lastRemainBuf.Write((*buf)[processed:])
				} else {
					// valid post
					remainLen := uint64(len(*buf) - processed)
					if remainLen == waitToBodyLen {
						postHttpRequest.Body = (*buf)[processed:]
						resultCh <- postHttpRequest.init()
						lastRemainBuf.Reset()
						lastInvalid = false
					} else if remainLen >= waitToBodyLen {
						postHttpRequest.Body = (*buf)[processed : processed+int(waitToBodyLen)]
						resultCh <- postHttpRequest.init()
						lastRemainBuf.Reset()
						lastRemainBuf.Write((*buf)[processed+int(waitToBodyLen):])
						lastInvalid = false
					} else {
						// wait Body
						waitPostBody = true
						lastRemainBuf.Reset()
						lastRemainBuf.Write((*buf)[processed:])
						waitPostBodyHttpRequest = postHttpRequest
					}
				}

			}
		} else {
			// process error, clear status
			waitPostBody = false
			lastRemainBuf.Reset()
			lastRemainBuf.Write(*buf)
			lastInvalid = true
		}
	}

}

type httpRequestListNode struct {
	value *HttpRequest
	next  *httpRequestListNode
}

type HttpRequestLinkedList struct {
	current *httpRequestListNode
	mux     sync.Mutex
}

func (l *HttpRequestLinkedList) Read() *HttpRequest {
	l.mux.Lock()
	defer l.mux.Unlock()
	if l.current != nil {
		ret := l.current.value
		l.current = l.current.next
		return ret
	} else {
		return nil
	}
}

func (l *HttpRequestLinkedList) Reset() {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.current = nil
}

// ordered and blocking way to process request
func ApplyRequestLinkedList(list *HttpRequestLinkedList, resultCh chan *HttpRequest, quit chan bool) {
	go func() {

		for {
			select {
			case result := <-resultCh:
				// write to list
				list.mux.Lock()
				current := list.current
				for current != nil {
					current = current.next
				}
				current = &httpRequestListNode{value: result}
				list.mux.Unlock()
			case <-quit:
				return
			}
		}
	}()
}
