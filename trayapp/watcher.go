package trayapp

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/fsnotify/fsnotify"
)

func watch(broadcaster *shared.Broadcaster) {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.WithError(err).Error("Failed to create file watcher")
			return
		}
		defer watcher.Close()

		err = watcher.Add(shared.StatePath)
		if err != nil {
			log.WithError(err).WithField("path", shared.StatePath).Error("Failed to add state file to watcher")
			return
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("State file modified, updating...")
					broadcaster.Send()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.WithError(err).Error("File watcher error")
			}
		}
	}()
}
