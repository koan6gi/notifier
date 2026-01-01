package updater

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/koan6gi/notifier/internal/config"
	"github.com/koan6gi/notifier/internal/domain"
	"github.com/koan6gi/notifier/pkg/logger"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"go.uber.org/zap"
)

const (
	pointFromIDKey = "point_from_id"
	pointToIDKey   = "point_to_id"
	dateKey        = "date"
	directionIDKey = "direction_id"

	requestTimeout = 30 * time.Second
)

type Repository interface {
	AddResponse(items []domain.Item)
}

type Worker struct {
	request *config.RequestConfig
	repo    Repository
}

func New(cfg *config.RequestConfig, repo Repository) *Worker {
	return &Worker{
		repo:    repo,
		request: cfg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.request.Interval)

	w.update(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			w.update(ctx)
		}
	}
}

func (w *Worker) update(ctx context.Context) {
	lg, ok := logger.FromContext(ctx)
	if !ok {
		log.Fatal("update worker: failed to get logger from context")
		return
	}

	lg = &logger.Logger{Logger: lg.With(zap.String("component", "updater"))}

	lg.Info("updater started", zap.String("time", time.Now().Format(time.RFC3339)))

	u := url.URL{
		Scheme: "https",
		Host:   w.request.Host,
		Path:   w.request.Path,
	}

	q := u.Query()

	q.Add(pointFromIDKey, strconv.Itoa(w.request.PointFromID))
	q.Add(pointToIDKey, strconv.Itoa(w.request.PointToID))
	q.Add(dateKey, w.request.Date)
	q.Add(directionIDKey, strconv.Itoa(w.request.DirectionID))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		lg.Error("failed to make request", zap.Error(err))
		return
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "auth.strategy=local")
	req.Header.Add("Host", "7588.by")
	req.Header.Add("Referer", "https://7588.by/route/pinsk/minsk")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Add("sec-ch-ua", `"Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Add("sec-ch-ua-mobile", `?0`)
	req.Header.Add("sec-ch-ua-platform", `"Linux"`)
	req.Header.Add("source", `site`)

	client := &http.Client{Timeout: requestTimeout}

	resp, err := client.Do(req)
	if err != nil {
		lg.Error("failed to do request", zap.Error(err))
		return
	}

	if resp.StatusCode != http.StatusOK {
		lg.Error("response code not 200 OK", zap.Int("code", resp.StatusCode), zap.String("status", http.StatusText(resp.StatusCode)))
		return
	}

	data := resp.Body.(io.Reader)
	defer resp.Body.Close()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzreader, err := gzip.NewReader(data)
		if err != nil {
			lg.Error("failed to create gzreader", zap.Error(err))
			return
		}
		defer gzreader.Close()

		b, err := io.ReadAll(gzreader)
		if err != nil {
			lg.Error("failed to decompress data", zap.Error(err))
			return
		}

		data = bytes.NewReader(b)
	}

	items := make([]domain.Item, 0)

	if err := json.NewDecoder(data).Decode(&items); err != nil {
		lg.Error("failed to decode json response", zap.Error(err))
		return
	}

	w.repo.AddResponse(items)

	found := false

	startTime, err := time.Parse(time.DateTime, w.request.StartTime)
	if err != nil {
		lg.Error("failed to parse start time", zap.Error(err))
		return
	}

	endTime, err := time.Parse(time.DateTime, w.request.EndTime)
	if err != nil {
		lg.Error("failed to parse end time", zap.Error(err))
		return
	}

	for _, item := range items {
		departureTime, err := time.Parse(time.DateTime, item.DepartureTime)
		if err != nil {
			lg.Error("failed to parse departure time", zap.String("time", item.DepartureTime), zap.Error(err))
			return
		}

		c, ok := item.Count.(float64)
		if startTime.Before(departureTime) && departureTime.Before(endTime) && (!ok || c > 0) {
			lg.Info("item found", zap.String("departure_time", item.DepartureTime), zap.Any("count", item.Count), zap.String("price", item.Price))

			found = true
		}
	}

	if found {
		go w.playSound()
	}

	lg.Info("updater finished")
}

func (w *Worker) playSound() {
	f, err := os.Open(w.request.SignalPath)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
