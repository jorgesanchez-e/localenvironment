package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/dns"
	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/publicip"
	"github.com/jorgesanchez-e/localenvironment/ddns/internal/domain/ddns"

	log "github.com/sirupsen/logrus"
)

const (
	ipv4Type int = iota
	ipv6Type
)

type configDecoder interface {
	Decode(node string, item any) error
}

type timeConfig struct {
	Timeout int `yaml:"process-timeout-seconds"`
	Elapse  int `yaml:"check-every-seconds"`
}

type Service struct {
	fetcher ddns.Fetcher
	updater ddns.Updater
	times   timeConfig
}

func New(ctx context.Context, cnf configDecoder) (*Service, error) {
	dnsDriver, err := dns.New(ctx, cnf)
	if err != nil {
		return nil, err
	}

	times := timeConfig{}
	if err := cnf.Decode("ddns", &times); err != nil {
		return nil, err
	}

	times.Elapse = times.Elapse - times.Timeout

	if times.Timeout == 0 || times.Elapse == 0 || times.Elapse <= times.Timeout {
		return nil, fmt.Errorf("configured times error, timeout:%d, elapse:%d", times.Timeout, times.Elapse)
	}

	return &Service{
		fetcher: publicip.New(),
		updater: dnsDriver,
		times:   times,
	}, nil
}

func (s *Service) process(ctx context.Context) {
	defer func() {
		<-ctx.Done()
		log.Debug("ending the process function")
	}()

	var ipv4, ipv6 net.IP

	log.Debugf("process starting, getting ips")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go s.getIp(ctx, &ipv4, ipv4Type, wg)
	go s.getIp(ctx, &ipv6, ipv6Type, wg)

	wg.Wait()

	if ipv4 == nil && ipv6 == nil {
		return
	}

	if err := s.updater.Update(ctx, []net.IP{ipv4, ipv6}); err == nil {
		log.Debug("all changes were updated")
	} else {
		if err == dns.ErrNoUpdates {
			log.Debug("no changes are needed to update")
		} else {
			log.Errorf("unable to make changes, error: %s", err)
		}
	}
}

func (s *Service) getIp(ctx context.Context, ip *net.IP, ipType int, wg *sync.WaitGroup) {
	ctxIp, cancel := context.WithTimeout(ctx, time.Duration(s.times.Timeout)*time.Second/2)

	defer func() {
		wg.Done()
		cancel()
	}()

	var err error
	switch ipType {
	case ipv4Type:
		*ip, err = s.fetcher.IPv4(ctxIp)
	case ipv6Type:
		*ip, err = s.fetcher.IPv6(ctxIp)
	default:
		log.Error("invalid ip type passed to get the new ips")
		return
	}

	ipTypeMessage := ""
	if ipType == ipv4Type {
		ipTypeMessage = "ipv4"
	} else {
		ipTypeMessage = "ipv6"
	}

	if err != nil {
		log.Errorf("unable to get %s, err:%s", ipTypeMessage, err)
	} else {
		log.Debugf("got %s information [%s]", ipTypeMessage, ip.String())
	}
}

func (s *Service) Run(ctx context.Context) chan bool {
	done := make(chan bool, 1)
	sig := make(chan os.Signal, 1)

	go s.run(ctx, sig, done)

	return done
}

func (s *Service) run(ctx context.Context, sig chan os.Signal, done chan bool) {
MainLoop:
	for {
		nextIteration, iterCancel := context.WithTimeout(ctx, time.Duration(s.times.Elapse)*time.Second)
		select {
		case <-sig:
			iterCancel()
			done <- true
			break MainLoop
		case <-nextIteration.Done():
			iterCancel()
			ctxTout, tOutCancel := context.WithTimeout(ctx, time.Duration(s.times.Timeout)*time.Second)
			s.process(ctxTout)
			tOutCancel()
		}
	}

	done <- true
}
