package helpers

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func LogSwarmServices(t *testing.T, ctx context.Context) {
	log, err := ReadAllServicesLog(ctx)
	require.NoError(t, err)
	go PrintLog(log, os.Stdout)
}

type LogLine struct {
	ServiceName string
	IsError     bool
	Text        string
}

type Log = <-chan *LogLine

func PrintLog(log Log, w io.Writer) {
	for {
		l, ok := <-log
		if ok {
			prefix := ""
			if l.IsError {
				prefix = "ERROR"
			}
			_, err := fmt.Fprintln(w, l.ServiceName, ":", prefix, l.Text)
			if err != nil {
				fmt.Println("error printing log line", err)
			}
		} else {
			return
		}
	}
}

func closeCloser(c io.Closer, name string) {
	if err := c.Close(); err != nil {
		fmt.Println("error closing", name, err)
	}
}

func ReadAllServicesLog(ctx context.Context) (Log, error) {
	logsLock := sync.Mutex{}
	logs := make(map[string]Log)
	multiLog := multiplexLogs(ctx, &logsLock, logs)
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	go func() {
		defer closeCloser(cli, "docker client")
		services := getServicesTicker(ctx, cli)
		for s := range services {
			if _, ok := logs[s.ID]; !ok {
				log, err := ReadServiceLog(ctx, cli, s)
				if err == nil {
					logsLock.Lock()
					logs[s.ID] = log
					logsLock.Unlock()
				} else {
					fmt.Println("error reading service log", err)
				}
			}
		}
	}()
	return multiLog, nil
}

func getServicesTicker(ctx context.Context, cli *client.Client) <-chan swarm.Service {
	servicesChan := make(chan swarm.Service)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if ctx.Err() != nil {
				return
			}
			services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
			if err == nil {
				for _, s := range services {
					servicesChan <- s
				}
			} else {
				fmt.Println("error getting service list", err)
			}
		}
	}()
	return servicesChan
}

func multiplexLogs(ctx context.Context, logsLock *sync.Mutex, logs map[string]Log) chan *LogLine {
	multiLog := make(chan *LogLine)
	go func() {
		defer close(multiLog)
		for {
			if ctx.Err() != nil {
				return
			}
			logsLock.Lock()
			for k, v := range logs {
				select {
				case l, ok := <-v:
					if ok {
						multiLog <- l
					} else {
						fmt.Println("service", k, "log died")
						delete(logs, k)
					}
				default:
				}
			}
			logsLock.Unlock()
		}
	}()
	return multiLog
}

func writerLogPipe(ctx context.Context, serviceName string, isError bool) (io.WriteCloser, Log) {
	src := ioutils.NewBytesPipe()
	dst := make(chan *LogLine)
	scanner := bufio.NewScanner(src)
	go func() {
		defer close(dst)
		for scanner.Scan() {
			if ctx.Err() != nil {
				return
			}
			dst <- &LogLine{
				ServiceName: serviceName,
				IsError:     isError,
				Text:        scanner.Text(),
			}
		}
	}()
	return src, dst
}

func ReadServiceLog(ctx context.Context, cli *client.Client, service swarm.Service) (Log, error) {
	muxLogs, err := cli.ServiceLogs(ctx, service.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	})
	if err != nil {
		return nil, err
	}
	stdOut, stdOutLog := writerLogPipe(ctx, service.Spec.Name, false)
	stdErr, stdErrLog := writerLogPipe(ctx, service.Spec.Name, true)

	go func() {
		defer closeCloser(muxLogs, "muxLogs")
		defer closeCloser(stdOut, "stdOut")
		defer closeCloser(stdErr, "stdErr")
		fmt.Println("started reading logs from", service.Spec.Name)
		_, err = stdcopy.StdCopy(stdOut, stdErr, muxLogs)
		if err != nil && ctx.Err() == nil {
			fmt.Println("error reading", service.Spec.Name, err)
		} else {
			fmt.Println("stopped reading logs from", service.Spec.Name)
		}
	}()
	return merge(stdOutLog, stdErrLog), nil
}

func merge(log1, log2 Log) Log {
	out := make(chan *LogLine)
	var i int32
	atomic.StoreInt32(&i, 2)
	pipe := func(c Log) {
		for v := range c {
			out <- v
		}
		if atomic.AddInt32(&i, -1) == 0 {
			close(out)
		}
	}
	go pipe(log1)
	go pipe(log2)
	return out
}
