package crack

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"sync"
	"time"
)

func SSH(ctx context.Context, cancel context.CancelFunc, ip, user, passwd string, port, timeout int, ch <-chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		<-ch
	}()
	select {
	case <-ctx.Done():
		return
	default:
	}
	configssh := &ssh.ClientConfig{
		Timeout:         time.Duration(timeout) * time.Second, // ssh连接timeout时间
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	configssh.Auth = []ssh.AuthMethod{ssh.Password(passwd)}

	addr := fmt.Sprintf("%s:%d", ip, port)
	sshClient, err := ssh.Dial("tcp", addr, configssh)
	if err != nil {
		return
	}
	defer sshClient.Close()
	end(ip, user, passwd, port, "SSH")
	cancel()
}
