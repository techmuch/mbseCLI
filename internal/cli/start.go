package cli

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"mbsecli/internal/feedback"
	"mbsecli/internal/parser"
	"mbsecli/internal/server"
	"mbsecli/internal/watch"
	"mbsecli/web"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var (
		dir      string
		port     int
		debounce time.Duration
		dev      bool
		open     bool
	)

	cmd := &cobra.Command{
		Use:     "start [file-or-directory]",
		Aliases: []string{"serve"},
		Short:   "Watch a .sysml file (or directory of them) and start the visualizer",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := dir
			if len(args) == 1 {
				target = args[0]
			}
			if target == "" {
				target = "."
			}
			return runStart(target, port, debounce, dev, open)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "directory (or file) to watch (default: current directory)")
	cmd.Flags().IntVar(&port, "port", 4173, "port to serve the web UI and API on")
	cmd.Flags().DurationVar(&debounce, "debounce", 200*time.Millisecond, "file-change debounce window")
	cmd.Flags().BoolVar(&dev, "dev", false, "dev mode: don't serve embedded assets (use with `npm run dev` in web/)")
	cmd.Flags().BoolVarP(&open, "open", "o", false, "open the browser automatically once started")

	return cmd
}

func runStart(target string, port int, debounce time.Duration, dev bool, open bool) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", target, err)
	}
	watchDir := target
	var singleFile string
	if !info.IsDir() {
		watchDir = filepath.Dir(target)
		singleFile = target
	}

	var assets = webAssetsOrNil(dev)
	srv := server.New(assets)

	loadAndPublish := func(path string) {
		src, err := os.ReadFile(path)
		if err != nil {
			log.Printf("read %s: %v", path, err)
			return
		}
		graph := parser.Parse(path, src)

		store, err := feedback.Open(path)
		if err != nil {
			log.Printf("feedback store for %s: %v", path, err)
		} else {
			live := map[string]bool{}
			for fqn := range graph.ByFQN {
				live[fqn] = true
			}
			store.MarkOrphaned(live)
		}

		srv.SetGraph(graph)
		srv.SetFeedback(store)
		srv.PublishUpdate()
		log.Printf("parsed %s (%d elements)", path, len(graph.ByFQN))
	}

	// Initial load.
	if singleFile != "" {
		loadAndPublish(singleFile)
	} else if first := firstSysMLFile(watchDir); first != "" {
		loadAndPublish(first)
	} else {
		log.Printf("no .sysml files found in %s yet — waiting for one to appear", watchDir)
	}

	w, err := watch.New(watchDir, debounce)
	if err != nil {
		return fmt.Errorf("starting watcher: %w", err)
	}
	defer w.Close()

	go func() {
		for ev := range w.Events {
			if singleFile != "" && ev.Path != singleFile {
				continue
			}
			if ev.Op == "remove" {
				continue
			}
			loadAndPublish(ev.Path)
		}
	}()

	addr := fmt.Sprintf(":%d", port)
	url := fmt.Sprintf("http://localhost%s", addr)
	log.Printf("mbsecli starting on %s (watching %s)", url, watchDir)

	if open {
		go func() {
			time.Sleep(100 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("failed to open browser: %v", err)
			}
		}()
	}

	return http.ListenAndServe(addr, srv.Routes())
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, openbsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func firstSysMLFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sysml" {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func webAssetsOrNil(dev bool) fs.FS {
	if dev {
		return nil
	}
	sub, err := web.Assets()
	if err != nil {
		log.Printf("embedded web assets unavailable (%v) — run `make build-web` or pass --dev", err)
		return nil
	}
	return sub
}
