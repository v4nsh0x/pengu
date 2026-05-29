package interpreter

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/v4nsh0x/pengu/runtime"
)

func createNetModule() *runtime.Value {
	om := runtime.NewOrderedMap()

	// net.scan(host, ports) - Scan an array of ports on a host
	om.Set("scan", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_ARRAY {
			return nil, fmt.Errorf("net.scan() expects (host, [ports])")
		}

		host := args[0].Str
		timeout := 2 * time.Second
		if len(args) >= 3 && args[2].Type == runtime.VAL_NUMBER {
			timeout = time.Duration(args[2].Number) * time.Second
		}

		result := runtime.NewOrderedMap()
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Limit concurrency to 100 goroutines
		sem := make(chan struct{}, 100)

		for _, portVal := range args[1].Array {
			if portVal.Type != runtime.VAL_NUMBER {
				continue
			}
			port := int(portVal.Number)
			wg.Add(1)
			sem <- struct{}{}
			go func(p int) {
				defer wg.Done()
				defer func() { <-sem }()

				addr := fmt.Sprintf("%s:%d", host, p)
				conn, err := net.DialTimeout("tcp", addr, timeout)
				mu.Lock()
				if err != nil {
					result.Set(fmt.Sprintf("%d", p), runtime.NewString("closed"))
				} else {
					conn.Close()
					result.Set(fmt.Sprintf("%d", p), runtime.NewString("open"))
				}
				mu.Unlock()
			}(port)
		}
		wg.Wait()
		return runtime.NewObject(result), nil
	}))

	// net.scan_range(host, start, end) - Scan a range of ports
	om.Set("scan_range", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 3 || args[0].Type != runtime.VAL_STRING ||
			args[1].Type != runtime.VAL_NUMBER || args[2].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("net.scan_range() expects (host, start_port, end_port)")
		}

		host := args[0].Str
		start := int(args[1].Number)
		end := int(args[2].Number)
		timeout := 2 * time.Second
		if len(args) >= 4 && args[3].Type == runtime.VAL_NUMBER {
			timeout = time.Duration(args[3].Number) * time.Second
		}

		if end-start > 10000 {
			return nil, fmt.Errorf("net.scan_range() max range is 10000 ports")
		}

		openPorts := make([]*runtime.Value, 0)
		var mu sync.Mutex
		var wg sync.WaitGroup
		sem := make(chan struct{}, 100)

		for p := start; p <= end; p++ {
			wg.Add(1)
			sem <- struct{}{}
			go func(port int) {
				defer wg.Done()
				defer func() { <-sem }()

				addr := fmt.Sprintf("%s:%d", host, port)
				conn, err := net.DialTimeout("tcp", addr, timeout)
				if err == nil {
					conn.Close()
					mu.Lock()
					openPorts = append(openPorts, runtime.NewNumber(float64(port), true))
					mu.Unlock()
				}
			}(p)
		}
		wg.Wait()
		return runtime.NewArray(openPorts), nil
	}))

	// net.connect(host, port) - Create a TCP connection, returns a connection object
	om.Set("connect", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 2 || args[0].Type != runtime.VAL_STRING || args[1].Type != runtime.VAL_NUMBER {
			return nil, fmt.Errorf("net.connect() expects (host, port)")
		}

		timeout := 5 * time.Second
		if len(args) >= 3 && args[2].Type == runtime.VAL_NUMBER {
			timeout = time.Duration(args[2].Number) * time.Second
		}

		addr := fmt.Sprintf("%s:%d", args[0].Str, int(args[1].Number))
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return nil, fmt.Errorf("net.connect() failed: %v", err)
		}

		connObj := runtime.NewOrderedMap()

		// conn.send(data)
		connObj.Set("send", runtime.NewBuiltin(func(sendArgs []*runtime.Value) (*runtime.Value, error) {
			if len(sendArgs) != 1 || sendArgs[0].Type != runtime.VAL_STRING {
				return nil, fmt.Errorf("conn.send() expects a string")
			}
			_, err := conn.Write([]byte(sendArgs[0].Str))
			if err != nil {
				return nil, fmt.Errorf("conn.send() failed: %v", err)
			}
			return runtime.NewBool(true), nil
		}))

		// conn.recv(size)
		connObj.Set("recv", runtime.NewBuiltin(func(recvArgs []*runtime.Value) (*runtime.Value, error) {
			size := 4096
			if len(recvArgs) >= 1 && recvArgs[0].Type == runtime.VAL_NUMBER {
				size = int(recvArgs[0].Number)
			}
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			buf := make([]byte, size)
			n, err := conn.Read(buf)
			if err != nil {
				return runtime.NewString(""), nil
			}
			return runtime.NewString(string(buf[:n])), nil
		}))

		// conn.close()
		connObj.Set("close", runtime.NewBuiltin(func(closeArgs []*runtime.Value) (*runtime.Value, error) {
			conn.Close()
			return runtime.NewBool(true), nil
		}))

		return runtime.NewObject(connObj), nil
	}))

	// net.dns_lookup(hostname) - Resolve hostname to IP addresses
	om.Set("dns_lookup", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("net.dns_lookup() expects a string hostname")
		}
		ips, err := net.LookupHost(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("net.dns_lookup() failed: %v", err)
		}
		arr := make([]*runtime.Value, len(ips))
		for i, ip := range ips {
			arr[i] = runtime.NewString(ip)
		}
		return runtime.NewArray(arr), nil
	}))

	// net.reverse_dns(ip) - Reverse DNS lookup
	om.Set("reverse_dns", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("net.reverse_dns() expects a string IP")
		}
		names, err := net.LookupAddr(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("net.reverse_dns() failed: %v", err)
		}
		arr := make([]*runtime.Value, len(names))
		for i, name := range names {
			arr[i] = runtime.NewString(strings.TrimSuffix(name, "."))
		}
		return runtime.NewArray(arr), nil
	}))

	// net.lookup_mx(domain) - MX record lookup
	om.Set("lookup_mx", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("net.lookup_mx() expects a string domain")
		}
		mxs, err := net.LookupMX(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("net.lookup_mx() failed: %v", err)
		}
		arr := make([]*runtime.Value, len(mxs))
		for i, mx := range mxs {
			item := runtime.NewOrderedMap()
			item.Set("host", runtime.NewString(strings.TrimSuffix(mx.Host, ".")))
			item.Set("priority", runtime.NewNumber(float64(mx.Pref), true))
			arr[i] = runtime.NewObject(item)
		}
		return runtime.NewArray(arr), nil
	}))

	// net.lookup_ns(domain) - NS record lookup
	om.Set("lookup_ns", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("net.lookup_ns() expects a string domain")
		}
		nss, err := net.LookupNS(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("net.lookup_ns() failed: %v", err)
		}
		arr := make([]*runtime.Value, len(nss))
		for i, ns := range nss {
			arr[i] = runtime.NewString(strings.TrimSuffix(ns.Host, "."))
		}
		return runtime.NewArray(arr), nil
	}))

	// net.lookup_txt(domain) - TXT record lookup
	om.Set("lookup_txt", runtime.NewBuiltin(func(args []*runtime.Value) (*runtime.Value, error) {
		if len(args) != 1 || args[0].Type != runtime.VAL_STRING {
			return nil, fmt.Errorf("net.lookup_txt() expects a string domain")
		}
		txts, err := net.LookupTXT(args[0].Str)
		if err != nil {
			return nil, fmt.Errorf("net.lookup_txt() failed: %v", err)
		}
		arr := make([]*runtime.Value, len(txts))
		for i, txt := range txts {
			arr[i] = runtime.NewString(txt)
		}
		return runtime.NewArray(arr), nil
	}))

	return runtime.NewObject(om)
}
