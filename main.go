/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.bytebuilders.dev/cert-manager-webhook-ace/cloudflare"

	whapi "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	dnsutil "github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	ctx := context.Background()

	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		panic("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(groupName,
		&aceDNSProviderSolver{ctx: ctx},
	)
}

type aceDNSProviderSolver struct {
	client    *kubernetes.Clientset
	ctx       context.Context
	userAgent string
}

type aceDNSProviderConfig struct {
	// Email of the account, only required when using API key based authentication.
	Email string

	// BaseURL of Cloudflare api, only required for running custom api proxy
	BaseURL string

	// API key to use to authenticate with Cloudflare.
	// Note: using an API token to authenticate is now the recommended method
	// as it allows greater control of permissions.
	APIKey *core.SecretKeySelector

	// API token used to authenticate with Cloudflare.
	APIToken *core.SecretKeySelector
}

func (c *aceDNSProviderSolver) Initialize(kubeClientConfig *restclient.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.client = cl
	c.userAgent = kubeClientConfig.UserAgent
	return nil
}

func (c *aceDNSProviderSolver) Present(ch *whapi.ChallengeRequest) error {
	log.Infof("Presenting challenge for fqdn=%s zone=%s", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := c.newClientFromConfig(ch)
	if err != nil {
		log.Errorf("failed to get client from ChallengeRequest: %s", err)
		return err
	}

	return client.Present(strings.TrimSuffix(ch.ResolvedZone, "."), ch.ResolvedFQDN, ch.Key)
}

func (c *aceDNSProviderSolver) CleanUp(ch *whapi.ChallengeRequest) error {
	log.Infof("Cleaning up entry for fqdn=%s", ch.ResolvedFQDN)
	client, err := c.newClientFromConfig(ch)
	if err != nil {
		log.Errorf("failed to get client from ChallengeRequest: %s", err)
		return fmt.Errorf("failed to get client from ChallengeRequest: %w", err)
	}
	return client.CleanUp(strings.TrimSuffix(ch.ResolvedZone, "."), ch.ResolvedFQDN, ch.Key)
}

func (c *aceDNSProviderSolver) Name() string {
	return "ace"
}

func (c *aceDNSProviderSolver) newClientFromConfig(ch *whapi.ChallengeRequest) (*cloudflare.DNSProvider, error) {
	cfg, err := c.loadConfig(ch)
	if err != nil {
		return nil, err
	}

	log.Info("preparing to create ace provider")
	if cfg.APIKey != nil && cfg.APIToken != nil {
		return nil, fmt.Errorf("API key and API token secret references are both present")
	}

	var selector *core.SecretKeySelector
	if cfg.APIKey != nil {
		selector = cfg.APIKey
	} else {
		selector = cfg.APIToken
	}

	keyData, err := c.loadSecretData(selector, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	var apiKey, apiToken string
	if cfg.APIKey != nil {
		apiKey = string(keyData)
	} else {
		apiToken = string(keyData)
	}

	email := cfg.Email
	p, err := cloudflare.NewDNSProviderCredentials(cfg.BaseURL, email, apiKey, apiToken, dnsutil.RecursiveNameservers, c.userAgent)
	if err != nil {
		return nil, fmt.Errorf("error instantiating ace challenge solver: %s", err)
	}
	return p, nil
}

func (c *aceDNSProviderSolver) loadConfig(ch *whapi.ChallengeRequest) (*aceDNSProviderConfig, error) {
	cfg := &aceDNSProviderConfig{}
	if ch.Config == nil {
		return cfg, nil
	}

	if err := json.Unmarshal(ch.Config.Raw, cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func (c *aceDNSProviderSolver) loadSecretData(selector *core.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := c.client.CoreV1().Secrets(ns).Get(c.ctx, selector.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", ns+"/"+selector.Name)
	}

	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}

	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}
