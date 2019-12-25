package sbc

import (
	sippy_log "sippy/log"
	sippy_types "sippy/types"
	"sync"
)

type CallManager struct {
	config     *Config
	logger     sippy_log.ErrorLogger
	Sip_TM     sippy_types.SipTransactionManager
	Proxy      sippy_types.StatefulProxy
	ccmap      map[int64]*CallController
	ccmap_lock sync.Mutex
}

func (self *CallManager) Remove(ccid int64) {
	self.ccmap_lock.Lock()
	defer self.ccmap_lock.Unlock()
	delete(self.ccmap, ccid)
}

func (self *CallManager) Shutdown() {
	self.ccmap_lock.Lock()
	defer self.ccmap_lock.Unlock()
	for _, cc := range self.ccmap {
		//println(cc.String())
		cc.Shutdown()
	}
}

func NewCallManager(config *Config, logger sippy_log.ErrorLogger) *CallManager {
	return &CallManager{
		logger: logger,
		config: config,
		ccmap:  make(map[int64]*CallController),
	}
}

var next_cc_id chan int64

func init() {
	next_cc_id = make(chan int64)
	go func() {
		var id int64 = 1
		for {
			next_cc_id <- id
			id++
		}
	}()
}

func (self *CallManager) OnNewDialog(req sippy_types.SipRequest, tr sippy_types.ServerTransaction) (sippy_types.UA, sippy_types.RequestReceiver, sippy_types.SipResponse) {
	toBody, err := req.GetTo().GetBody(self.config)
	if err != nil {
		self.logger.Error("CallManager::OnNewDialog: #1: " + err.Error())
		return nil, nil, req.GenResponse(500, "Internal Server Error", nil, nil)
	}
	if toBody.GetTag() != "" {
		// Request within dialog, but no such dialog
		return nil, nil, req.GenResponse(481, "Call Leg/Transaction Does Not Exist", nil, nil)
	}
	if req.GetMethod() == "INVITE" {
		// New dialog
		cc := NewCallController(self, <-next_cc_id)
		self.ccmap_lock.Lock()
		self.ccmap[cc.id] = cc
		self.ccmap_lock.Unlock()
		return cc.uaA, cc.uaA, nil
	}
	if req.GetMethod() == "REGISTER" {
		// Registration
		return nil, self.Proxy, nil
	}
	if req.GetMethod() == "NOTIFY" || req.GetMethod() == "PING" {
		// Whynot?
		return nil, nil, req.GenResponse(200, "OK", nil, nil)
	}
	if req.GetMethod() == "OPTIONS" {
		return nil, nil, req.GenResponse(200, "OK", nil, nil)
	}
	return nil, nil, req.GenResponse(501, "Not Implemented", nil, nil)
}
