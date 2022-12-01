package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/go-yaml/yaml"

	webhookconfig "g.hz.netease.com/horizon/pkg/config/webhook"
	eventmanager "g.hz.netease.com/horizon/pkg/event/manager"
	"g.hz.netease.com/horizon/pkg/eventhandler/wlgenerator"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/common"
	"g.hz.netease.com/horizon/pkg/util/log"
	webhookmanager "g.hz.netease.com/horizon/pkg/webhook/manager"
	"g.hz.netease.com/horizon/pkg/webhook/models"
	webhookmodels "g.hz.netease.com/horizon/pkg/webhook/models"
)

type worker struct {
	idleWaitInterval         uint
	responseBodyTruncateSize uint

	ctx            context.Context
	insecureClient http.Client
	secureClient   http.Client
	quit           chan bool
	webhook        *models.Webhook

	webhookManager webhookmanager.Manager
	eventManager   eventmanager.Manager
	userManager    usermanager.Manager
}

type Service interface {
	Start()
	StopAndWait()
}

type service struct {
	config  webhookconfig.Config
	ctx     context.Context
	quit    chan bool
	workers map[uint]*worker

	webhookManager webhookmanager.Manager
	eventManager   eventmanager.Manager
	userManager    usermanager.Manager
}

func NewService(ctx context.Context, manager *managerparam.Manager, config webhookconfig.Config) Service {
	return &service{
		config:         config,
		ctx:            ctx,
		workers:        make(map[uint]*worker),
		webhookManager: manager.WebhookManager,
		eventManager:   manager.EventManager,
		userManager:    manager.UserManager,
	}
}

func (s *service) stopWorkersAndWait() {
	wg := sync.WaitGroup{}
	wg.Add(len(s.workers))
	for _, w := range s.workers {
		go func(wk *worker) {
			wk.Stop().Wait()
			wg.Done()
		}(w)
	}
	wg.Wait()
}

func (s *service) reconcileWorkers() {
	// 1. get latest webhook list
	webhooks, err := s.webhookManager.ListWebhooks(s.ctx)
	if err != nil {
		log.Errorf(s.ctx, "failed to list webhooks, error: %s", err.Error())
		return
	}
	// 2. compare and reconcile workers
	reconciled := map[uint]bool{}
	for _, webhook := range webhooks {
		id := webhook.ID
		if worker, ok := s.workers[id]; ok {
			// 2.1 update workers
			if worker.webhook.UpdatedAt.Before(webhook.UpdatedAt) {
				worker.webhook = webhook
			}
		} else {
			// 2.2 create workers
			s.workers[id] = newWebhookWorker(s.webhookManager, s.eventManager,
				s.userManager, webhook, s.config.IdleWaitInterval,
				s.config.ClientTimeout, s.config.ResponseBodyTruncateSize)
		}
		reconciled[id] = true
	}

	// 2.2 stop deleted workers
	// todo: stop disabled worker
	for id := range s.workers {
		if _, ok := reconciled[id]; ok {
			continue
		}
		s.workers[id].Stop()
		delete(s.workers, id)
	}
}

// start webhook service to reconcile workers
func (s *service) Start() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf(s.ctx, "webhook service reconcile panic: %v", err)
				common.PrintStack()
			}
		}()

		t := time.NewTicker(time.Second * time.Duration(s.config.WorkerReconcileInterval))
		s.reconcileWorkers()
	L:
		for {
			select {
			case <-s.quit:
				s.stopWorkersAndWait()
				close(s.quit)
				break L
			case <-t.C:
				s.reconcileWorkers()
			}
		}
	}()
}

func (s *service) StopAndWait() {
	s.quit <- true
	<-s.quit
	log.Info(s.ctx, "webhook service stopped")
}

func newWebhookWorker(webhookMgr webhookmanager.Manager,
	eventMgr eventmanager.Manager, userMgr usermanager.Manager,
	webhook *models.Webhook, idleWaitInterval uint,
	clientTimeout, responseBodyTruncateSize uint) *worker {
	ww := &worker{
		idleWaitInterval:         idleWaitInterval,
		responseBodyTruncateSize: responseBodyTruncateSize,
		ctx:                      context.Background(),
		webhook:                  webhook,
		quit:                     make(chan bool, 1),
		insecureClient: http.Client{
			Timeout: time.Second * time.Duration(clientTimeout),
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		secureClient: http.Client{
			Timeout: time.Second * time.Duration(clientTimeout),
		},
		webhookManager: webhookMgr,
		eventManager:   eventMgr,
		userManager:    userMgr,
	}
	go ww.start()
	return ww
}

func (w *worker) sendWebhook(ctx context.Context, wl *models.WebhookLog) {
	saveResult := func() {
		if wl.ErrorMessage != "" {
			wl.Status = webhookmodels.StatusFailed
		} else {
			wl.Status = webhookmodels.StatusSuccess
		}
		_, err := w.webhookManager.UpdateWebhookLog(ctx, wl)
		if err != nil {
			log.Errorf(ctx, "failed to update webhook log %d, error: %s", wl.ID, err.Error())
		}
	}

	defer saveResult()

	// 1. make request and set body
	reqBody, err := addWebhookLogID([]byte(wl.RequestData), wl.ID)
	if err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to add id, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		return
	}
	req, err := http.NewRequest(http.MethodPost, wl.URL,
		bytes.NewBuffer(reqBody))
	if err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to new request, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		return
	}

	// 2. set headers
	headers := http.Header{}
	if err := yaml.Unmarshal([]byte(wl.RequestHeaders), &headers); err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to unmarshal header, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		return
	}
	req.Header = headers

	// 3. send request
	cli := w.secureClient
	if !w.webhook.SSLVerifyEnabled {
		cli = w.insecureClient
	}
	resp, err := cli.Do(req)
	if err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to send req, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		return
	}

	// 4. update response body
	respBody, err := ioutil.ReadAll(
		io.LimitReader(resp.Body, int64(w.responseBodyTruncateSize)),
	)
	if err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to read response body, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		resp.Body.Close()
		return
	}
	wl.ResponseBody = string(respBody)

	// 5. update response headers
	respHeader, err := yaml.Marshal(resp.Header)
	if err != nil {
		wl.ErrorMessage = fmt.Sprintf("failed to marshal, error: %+v", err)
		log.Errorf(ctx, wl.ErrorMessage)
		resp.Body.Close()
		return
	}
	if resp.StatusCode >= http.StatusBadRequest || resp.StatusCode < http.StatusOK {
		wl.ErrorMessage = fmt.Sprintf("unexpected response code: %d", resp.StatusCode)
	}
	wl.ResponseHeaders = string(respHeader)
	resp.Body.Close()
}

// start webhook worker and begin to send process webhook logs
func (w *worker) start() {
	ctx := w.ctx
	defer func() {
		if err := recover(); err != nil {
			log.Errorf(ctx, "webhook worker panic: %v", err)
			common.PrintStack()
		}
	}()
L:
	for {
		select {
		case <-w.quit:
			close(w.quit)
			break L
		default:
			// todo: set limit and find a way to avoid this invoke
			wls, err := w.webhookManager.ListWebhookLogsByStatus(ctx, w.webhook.ID,
				webhookmodels.StatusWaiting)
			if err != nil {
				log.Errorf(ctx, "failed to list webhook logs of %d, error: %s", w.webhook.ID, err.Error())
				continue
			}
			if len(wls) == 0 {
				time.Sleep(time.Second * time.Duration(w.idleWaitInterval))
				continue
			}
			for _, wl := range wls {
				w.sendWebhook(ctx, wl)
			}
		}
	}
}

func (w *worker) Stop() *worker {
	w.quit <- true
	return w
}

func (w *worker) Wait() {
	<-w.quit
	log.Infof(w.ctx, "webhook worker %d stopped", w.webhook.ID)
}

func addWebhookLogID(reqData []byte, id uint) ([]byte, error) {
	var content wlgenerator.MessageContent
	err := json.Unmarshal([]byte(reqData), &content)
	if err != nil {
		return nil, err
	}

	content.ID = id
	return json.Marshal(content)
}
