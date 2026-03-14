package sftpd

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/pidrive/pidrive/internal/auth"
)

type SFTPServer struct {
	authService *auth.AuthService
	mountPath   string
	hostKeyPath string
	addr        string
}

func NewSFTPServer(authSvc *auth.AuthService, mountPath, hostKeyPath, addr string) *SFTPServer {
	return &SFTPServer{
		authService: authSvc,
		mountPath:   mountPath,
		hostKeyPath: hostKeyPath,
		addr:        addr,
	}
}

func (s *SFTPServer) Start() error {
	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			apiKey := string(password)
			agent, err := s.authService.Authenticate(apiKey)
			if err != nil {
				return nil, fmt.Errorf("auth failed")
			}
			if !agent.Verified {
				return nil, fmt.Errorf("agent not verified")
			}
			return &ssh.Permissions{
				Extensions: map[string]string{
					"agent_id": agent.ID,
					"email":    agent.Email,
				},
			}, nil
		},
	}

	hostKey, err := s.loadOrGenerateHostKey()
	if err != nil {
		return fmt.Errorf("host key error: %w", err)
	}
	sshConfig.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	log.Printf("[sftp] listening on %s", s.addr)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("[sftp] accept error: %v", err)
				continue
			}
			go s.handleConnection(conn, sshConfig)
		}
	}()

	return nil
}

func (s *SFTPServer) handleConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	agentID := sshConn.Permissions.Extensions["agent_id"]
	email := sshConn.Permissions.Extensions["email"]
	log.Printf("[sftp] agent connected: %s (%s)", email, agentID)

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go func(in <-chan *ssh.Request) {
			for req := range in {
				if req.Type == "subsystem" && len(req.Payload) > 4 {
					subsystem := string(req.Payload[4:])
					if subsystem == "sftp" {
						req.Reply(true, nil)
						continue
					}
				}
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}(requests)

		agentRoot := filepath.Join(s.mountPath, "agents", agentID, "files")
		os.MkdirAll(agentRoot, 0755)

		// Use a chrooted filesystem handler
		handler := sftp.Handlers{
			FileGet:  &chrootHandler{root: agentRoot},
			FilePut:  &chrootHandler{root: agentRoot},
			FileCmd:  &chrootHandler{root: agentRoot},
			FileList: &chrootHandler{root: agentRoot},
		}

		server := sftp.NewRequestServer(channel, handler)
		if err := server.Serve(); err != nil {
			if err != io.EOF {
				log.Printf("[sftp] serve error for %s: %v", email, err)
			}
		}
		server.Close()
		log.Printf("[sftp] agent disconnected: %s", email)
	}
}

func (s *SFTPServer) loadOrGenerateHostKey() (ssh.Signer, error) {
	if data, err := os.ReadFile(s.hostKeyPath); err == nil {
		return ssh.ParsePrivateKey(data)
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}

	pemBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}

	os.MkdirAll(filepath.Dir(s.hostKeyPath), 0700)
	f, err := os.OpenFile(s.hostKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	pem.Encode(f, pemBlock)
	f.Close()

	return ssh.ParsePrivateKey(pem.EncodeToMemory(pemBlock))
}
