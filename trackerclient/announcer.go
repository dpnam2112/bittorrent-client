package tracker

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dpnam2112/bittorrent-client/common"
)

// events from TrackerAnnouncer
type AnnouncerEvent struct {
	AnnounceResp *AnnounceResponse
	Err error
}

// TrackerAnnouncer resolves peers by sending announcement requests to tracker(s).
// For each announcement request, trackers only returns a subset of peers. It's the
// responsibility of the user (caller) to manage peer connections and to track which peers are
// already connected to, which are not.
type TrackerAnnouncer interface {
	// announce request to be used when running RunAutoAnnouncer
	SetAnnounceRequest(AnnounceRequest)
	AnnounceRequest() AnnounceRequest

	// can be used as a goroutine to automatically send annoucement to a tracker after a
	// specific interval. the interval between two announcements for a tracker is based on the
	// field 'interval' in  annoucement responses from that tracker.
	RunAutoAnnouncer(context.Context) error
	Events() <-chan AnnouncerEvent 
}

func NewTrackerAnnouncer(ctx context.Context, metainfo *common.TorrentMetainfo, logger slog.Logger) (TrackerAnnouncer, error) {
	announcer := trackerAnnouncerImpl{
		logger: logger,
		announcerEvents: make(chan AnnouncerEvent, 16),
	}

	trackerAddrs := []string{metainfo.Announce()}

	for _, announceList := range metainfo.AnnounceList() {
		trackerAddrs = append(trackerAddrs, announceList...)
	}

	for i, addr := range trackerAddrs {
		trackerClient, err := NewTrackerClientFromURL(ctx, addr, logger)
		if err != nil && i < len(trackerAddrs) - 1 {
			continue
		}

		if err != nil && i == len(trackerAddrs) - 1 {
			return nil, fmt.Errorf("Cannot create tracker from announce and announce list")
		}

		announcer.trackerClient = trackerClient
		return &announcer, nil
	}

	return nil, fmt.Errorf("NewTrackerAnnouncer: Unknown error. Maybe there are no trackers specified in the torrent.")
}

type trackerAnnouncerImpl struct {
	req AnnounceRequest
	reqRwLock sync.RWMutex
	trackerClient TrackerClient
	logger slog.Logger
	announcerEvents chan AnnouncerEvent
}

func (trackerAnnouncer *trackerAnnouncerImpl) SetAnnounceRequest(req AnnounceRequest) {
	trackerAnnouncer.reqRwLock.Lock()
	defer trackerAnnouncer.reqRwLock.Unlock()
	trackerAnnouncer.req = req
}

func (trackerAnnouncer *trackerAnnouncerImpl) AnnounceRequest() AnnounceRequest {
	trackerAnnouncer.reqRwLock.RLock()
	defer trackerAnnouncer.reqRwLock.RUnlock()
	return trackerAnnouncer.req
}

func (trackerAnnouncer *trackerAnnouncerImpl) Events() <-chan AnnouncerEvent {
	return trackerAnnouncer.announcerEvents
}

// this method is expected to be executed as a goroutine
func (trackerAnnouncer  *trackerAnnouncerImpl) RunAutoAnnouncer(ctx context.Context) error {
	if trackerAnnouncer.trackerClient == nil {
		return fmt.Errorf("RunAutoAnnouncer: tracker client is nil")
	}

	waitingTime := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(waitingTime) * time.Second):
			resp, err := trackerAnnouncer.trackerClient.Announce(ctx, trackerAnnouncer.AnnounceRequest())
			if err != nil {
				trackerAnnouncer.announcerEvents <- AnnouncerEvent{AnnounceResp: nil, Err: err}
			} else {
				trackerAnnouncer.announcerEvents <- AnnouncerEvent{AnnounceResp: &resp, Err: err}
				waitingTime = int(resp.Interval)
			}
		}
	}
}

func parseTrackerURL(trackerURL string) (host string, port int, scheme string, err error) {
	// Parse the URL
	u, err := url.Parse(trackerURL)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid URL: %w", err)
	}

	// Extract host and port from u.Host
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		// If port is missing, try default per scheme
		host = u.Host
		switch u.Scheme {
		case "http":
			portStr = "80" // default to 80 even for https, or customize
		case "https":
			portStr = "443"
		case "udp":
			portStr = "80"
		default:
			return "", 0, u.Scheme, fmt.Errorf("unknown scheme: %s", u.Scheme)
		}
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, u.Scheme, fmt.Errorf("invalid port: %w", err)
	}

	return host, portInt, u.Scheme, nil
}

func resolveHostToIPs(host string) ([]net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("DNS lookup failed: %w", err)
	}
	var v4, v6 []net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			v4 = append(v4, ip)
		} else {
			v6 = append(v6, ip)
		}
	}
	if len(v4) > 0 {
		return v4, nil
	}
	return v6, nil
}
