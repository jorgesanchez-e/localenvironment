package aws

import (
	"context"
	"crypto"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/providers/dns/route53"
	"github.com/go-acme/lego/v5/registration"
	"github.com/jorgesanchez-e/localenvironment/apps/acme/internal/domain/certs"
)

const (
	staging bool = true
)

var (
	ErrParsingPrivateKey       = errors.New("aws-lego: parsing primary key")
	ErrAcmeLegoClient          = errors.New("aws-lego: creating acme lego client")
	ErrRoute53ProviderCreation = errors.New("aws-lego: route53 provider creation")
	ErrAcmeAccountRegistration = errors.New("aws-lego: account registration")
	ErrInvalidCertificate      = errors.New("aws-lego: invalid certificate resource")
	ErrAcmeGettingCertificate  = errors.New("aws-lego: getting acme certificate")
	ErrSettingCerticiate       = errors.New("setting certificate")
	ErrSettingPK               = errors.New("setting private key")
	ErrSettingIssuer           = errors.New("setting issuer")
	ErrSettingCSR              = errors.New("setting csr")
)

// newClientBuilder is a package variable so tests can replace the real lego client builder.
var newClientBuilder = legoClientBuilder

type acmeClient interface {
	SetDNS01Provider(p challenge.Provider, opts ...dns01.ChallengeOption) error
	Register(ctx context.Context, opts registration.RegisterOptions) (*acme.ExtendedAccount, error)
	Obtain(ctx context.Context, req certificate.ObtainRequest) (*certificate.Resource, error)
	Renew(ctx context.Context, certRes certificate.Resource, options *certificate.RenewOptions) (*certificate.Resource, error)
}

type Driver struct {
	clientBuilder func() (acmeClient, error)
	user          acmeUser
	awsAccounts   []awsConfig
}

type acmeUser struct {
	Email   string
	Account *acme.ExtendedAccount
	key     crypto.Signer
}

func (u *acmeUser) GetEmail() string                       { return u.Email }
func (u *acmeUser) GetRegistration() *acme.ExtendedAccount { return u.Account }
func (u *acmeUser) GetPrivateKey() crypto.Signer           { return u.key }

func New(ctx context.Context, cnf Configurator) (*Driver, error) {
	config, err := readConfig(cnf)
	if err != nil {
		return nil, err
	}

	primaryKey, err := certcrypto.ParsePEMPrivateKey([]byte(config.PrimaryKey))
	if err != nil {
		return nil, fmt.Errorf("%w error:%q", ErrParsingPrivateKey, err)
	}

	driver := new(Driver)
	driver.user.Email = config.Email
	driver.user.key = primaryKey

	driver.awsAccounts = make([]awsConfig, len(config.AWSConf))
	for index, acct := range config.AWSConf {
		driver.awsAccounts[index].Accesskey = acct.Accesskey
		driver.awsAccounts[index].SecretKey = acct.SecretKey
		driver.awsAccounts[index].Region = acct.Region
		driver.awsAccounts[index].Domains = make([]route53DomainConfig, len(acct.Domains))
		copy(driver.awsAccounts[index].Domains, acct.Domains)
	}

	tempClient, err := newClientBuilder(&driver.user)()
	if err != nil {
		return nil, fmt.Errorf("%w, unable to create temporary client, error:%q", ErrAcmeLegoClient, err)
	}

	acmeAccount, err := tempClient.Register(ctx, registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("%w, unable to register acme user, error:%q", ErrAcmeAccountRegistration, err)
	}

	driver.user = acmeUser{
		Email:   config.Email,
		Account: acmeAccount,
		key:     primaryKey,
	}
	driver.clientBuilder = newClientBuilder(&driver.user)

	return driver, nil
}

func (d *Driver) Upsert(ctx context.Context, certificates []*certs.Cert) error {
	wg := new(sync.WaitGroup)
	errs := make([]error, len(certificates))
	wg.Add(len(certificates))

	for index, c := range certificates {
		go func() {
			defer wg.Done()
			dns01.SetDefaultClient(dns01.NewClient(&dns01.Options{
				RecursiveNameservers: []string{"8.8.8.8:53", "1.1.1.1:53"},
			}))

			var err error
			var resource *certificate.Resource

			if resource, err = d.getCertificate(ctx, c); err != nil {
				errs[index] = err
				return
			}

			if err = setCertificateData(certificates[index], resource); err != nil {
				errs[index] = err
			}
		}()
	}

	wg.Wait()

	return errors.Join(errs...)
}

func (d *Driver) Revoke(ctx context.Context, cert []*certs.Cert) error {
	return nil
}

func (d *Driver) getCertificate(ctx context.Context, c *certs.Cert) (*certificate.Resource, error) {
	var client acmeClient
	var credentials *awsCredentials
	var err error

	if client, err = d.clientBuilder(); err != nil {
		return nil, fmt.Errorf("%w, error:%q", ErrAcmeLegoClient, err)
	}

	for _, account := range d.awsAccounts {
		if creds, err := account.getCredentials(c.Domain()); err != nil {
			continue
		} else {
			credentials = new(awsCredentials)
			credentials = creds
		}
	}

	if credentials == nil {
		return nil, fmt.Errorf("%w, domain:%q", ErrAWSCredentialsNotFound, c.Domain())
	}

	r53cfg := route53.NewDefaultConfig()
	r53cfg.AccessKeyID = credentials.accessKey
	r53cfg.SecretAccessKey = credentials.secretKey
	r53cfg.Region = credentials.region
	r53cfg.HostedZoneID = credentials.zoneID
	r53cfg.PropagationTimeout = 5 * time.Minute
	r53cfg.PollingInterval = 10 * time.Second

	var provider *route53.DNSProvider
	provider, err = route53.NewDNSProviderConfig(r53cfg)
	if err != nil {
		return nil, fmt.Errorf("%w error: %w", ErrRoute53ProviderCreation, err)
	}

	if err = client.SetDNS01Provider(provider); err != nil {
		return nil, fmt.Errorf("configurando DNS-01: %w", err)
	}

	var resource *certificate.Resource
	resource, err = client.Obtain(ctx, certificate.ObtainRequest{
		Domains: []string{c.Domain()},
		Bundle:  true,
		KeyType: certcrypto.EC256,
	})

	if err != nil {
		return nil, fmt.Errorf("%w, error:%q", ErrAcmeGettingCertificate, err)
	}

	return resource, nil
}

func setCertificateData(c *certs.Cert, resource *certificate.Resource) error {
	var err error

	if resource == nil {
		return fmt.Errorf("%w error:%q", ErrInvalidCertificate, err)
	}

	if err = c.SetCertificate(resource.Certificate); err != nil {
		return fmt.Errorf("%w error:%q", ErrSettingCerticiate, err)
	}

	if err = c.SetPrivateKey(resource.PrivateKey); err != nil {
		return fmt.Errorf("%w error:%q", ErrSettingPK, err)
	}

	if err = c.SetIssuerCertificate(resource.IssuerCertificate); err != nil {
		return fmt.Errorf("%w error:%q", ErrSettingIssuer, err)
	}

	if err = c.SetCSR(resource.CSR); err != nil {
		return fmt.Errorf("%w error:%q", ErrSettingCSR, err)
	}

	return nil
}
